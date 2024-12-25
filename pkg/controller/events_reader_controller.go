package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/sink"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"
)

type EventController struct {
	eventInformer cache.Controller

	eventIndexer cache.Store

	queue workqueue.TypedRateLimitingInterface[KeyEvent]
	sinks []sink.ISink
}

type KeyEvent struct {
	Key       string
	EventType watch.EventType
}

// NewClusterEventController creates a new *EventController that will watch
// for k8s.io/api/core/v1 Event resources in all namespaces
func NewClusterEventController(clientSet kubernetes.Interface, newListerWatcherFunc func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher, sinks []sink.ISink) *EventController {
	return newEventController(clientSet, v1.NamespaceAll, newListerWatcherFunc, sinks)
}

// NewNamespacedEventControllers creates an array of *EventController type that will watch
// for k8s.io/api/core/v1 Event resources only in the set of namespaces
func NewNamespacedEventControllers(clientSet kubernetes.Interface, namespaces []string, newListerWatcherFunc func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher, sinks []sink.ISink) []*EventController {
	eventControllers := make([]*EventController, len(namespaces))

	for i, namespace := range namespaces {
		eventController := newEventController(clientSet, namespace, newListerWatcherFunc, sinks)
		eventControllers[i] = eventController
	}

	return eventControllers
}

// newEventController implements inner creation of EventController instance
func newEventController(clientSet kubernetes.Interface, namespace string, newListerWatcherFunc func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher, sinks []sink.ISink) *EventController {
	if sinks == nil || len(sinks) < 1 {
		return nil
	}
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[KeyEvent]()
	queue := workqueue.NewTypedRateLimitingQueue(rateLimiter)
	indexer, informer := NewIndexerInformer(clientSet.CoreV1().RESTClient(), namespace, queue, newListerWatcherFunc)

	return &EventController{
		queue:         queue,
		eventInformer: informer,
		eventIndexer:  indexer,
		sinks:         sinks,
	}
}

// NewListerWatcherFunc returns function to create cache.ListerWatcher
func NewListerWatcherFunc() func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher {
	return func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher {
		return newListWatchFromClient(kubeRestClient, "events", namespace, fields.Everything())
	}
}

// newListWatchFromClient creates cache.ListWatch with empty ListFunc to avoid overfilling cache ob objects during startup
func newListWatchFromClient(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector) cache.ListerWatcher {
	optionsModifier := func(options *v1.ListOptions) {
		options.FieldSelector = fieldSelector.String()
	}
	listFunc := func(options v1.ListOptions) (runtime.Object, error) {
		return &corev1.EventList{}, nil
	}
	watchFunc := func(options v1.ListOptions) (watch.Interface, error) {
		options.Watch = true
		optionsModifier(&options)
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			Throttle(getThrottleTokenBucketRateLimiter()).
			VersionedParams(&options, v1.ParameterCodec).
			Watch(context.TODO())
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

// getThrottleTokenBucketRateLimiter creates token bucket ratelimiter with parameters from environment variables
func getThrottleTokenBucketRateLimiter() flowcontrol.RateLimiter {
	qps := os.Getenv("WATCH_QPS")
	if len(qps) < 1 {
		qps = "5"
	}
	qpsV, err := strconv.Atoi(qps)
	if err != nil {
		qpsV = 5
	}
	burst := os.Getenv("WATCH_BURST")
	if len(burst) < 1 {
		burst = "10"
	}
	burstV, err := strconv.Atoi(burst)
	if err != nil {
		burstV = 10
	}
	slog.Debug("flowcontrol.tokenBucketRateLimiter will be used for throttling watch requests with parameters", "WATCH_QPS", qpsV, "WATCH_BURST", burstV)
	return flowcontrol.NewTokenBucketRateLimiter(float32(qpsV), burstV)
}

// NewIndexerInformer returns newly created indexer and informer for watcher
func NewIndexerInformer(kubeRestClient rest.Interface, namespace string, queue workqueue.TypedRateLimitingInterface[KeyEvent], watcherFunc func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher) (cache.Store, cache.Controller) {
	eventListWatcher := watcherFunc(kubeRestClient, namespace)

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(event interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(event)
			//todo here can be added some filters to not add to queue
			if err == nil {
				queue.AddRateLimited(KeyEvent{Key: key, EventType: watch.Added})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newEvent := newObj.(*corev1.Event)
			oldEvent := oldObj.(*corev1.Event)
			if newEvent.ResourceVersion == oldEvent.ResourceVersion {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.AddRateLimited(KeyEvent{Key: key, EventType: watch.Modified})
			}
		},
	}

	options := cache.InformerOptions{
		ListerWatcher: eventListWatcher,
		ObjectType:    &corev1.Event{},
		Handler:       handlers,
		ResyncPeriod:  0,
		Indexers:      cache.Indexers{},
	}

	indexer, informer := cache.NewInformerWithOptions(options)
	return indexer, informer
}

// Run starts workers for syncing events with eventInformer
func (c *EventController) Run(workers int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	go c.eventInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.eventInformer.HasSynced) {
		slog.Error("failed to wait for caches to sync")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	slog.Info("started workers")
	<-stopCh
	slog.Info("shutting down workers")
}

// runWorker constantly processes each event
func (c *EventController) runWorker() {
	for c.processNextWorkItem() {
	}
}

var nilKeyEvent = KeyEvent{}

// processNextWorkItem will read each item from queue and
// attempt to process it, by calling the syncHandler
func (c *EventController) processNextWorkItem() bool {
	keyEvent, shutdown := c.queue.Get()

	if shutdown {
		slog.Debug("Shutdown signal is called")
		return false
	}
	defer c.queue.Done(keyEvent)

	var err error
	if keyEvent != nilKeyEvent {
		if keyEvent.EventType == watch.Added {
			slog.Debug("add event is triggered for object", "key", keyEvent.Key)
		} else if keyEvent.EventType == watch.Modified {
			slog.Debug("modify event is triggered for object", "key", keyEvent.Key)
		}
		err = c.syncHandler(keyEvent.Key)
	}

	c.handleErr(err, keyEvent)

	return true
}

// syncHandler gets item from eventIndexer by key and call processEvent
func (c *EventController) syncHandler(key string) error {
	obj, exists, err := c.eventIndexer.GetByKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}
	if !exists {
		return nil
	}
	err = c.processEvent(obj)
	if err == nil {
		//clearing store after processing event immediately
		err = c.eventIndexer.Delete(obj)
		if err != nil {
			slog.Error("Failed to delete event from store")
		}
	}
	return err
}

// processEvent implements logic of processing event and printing it to stdout
func (c *EventController) processEvent(obj interface{}) error {

	slog.Debug("process triggered for an object", "object", obj)
	eventObj, ok := obj.(*corev1.Event)
	if !ok {
		err := fmt.Errorf("could not convert object to v1.Event type")
		slog.Error(err.Error())
		return err
	}

	var joinedErr error
	for _, s := range c.sinks {
		if err := s.Release(eventObj); err != nil {
			joinedErr = errors.Join(joinedErr, err)
		}
	}
	return joinedErr
}

// handleErr checks if an error happened and makes attempts to reprocess item
func (c *EventController) handleErr(err error, key KeyEvent) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 3 {
		slog.Error("error syncing event", "error", err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	slog.Info("dropping event out of the queue with error", "error", err)
}
