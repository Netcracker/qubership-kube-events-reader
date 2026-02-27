package controller

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/format"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/sink"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	fcache "k8s.io/client-go/tools/cache/testing"
)

var fKubeClient = fake.NewClientset()

// FakeListerWatcherFunc returns function to create cache.ListerWatcher for testing
func FakeListerWatcherFunc(source *fcache.FakeControllerSource) func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher {
	return func(kubeRestClient rest.Interface, namespace string) cache.ListerWatcher {
		return source
	}
}

func Test_ClusterEventController_StdoutSink(t *testing.T) {
	var fakeLW = fcache.NewFakeControllerSource()
	// change stdout to print events in file
	initialStdout := os.Stdout
	fname, err := test.ChangeStdoutToFile("stdout1")
	defer func(t *testing.T) {
		assert.NoError(t, test.ChangeFileToStdout(initialStdout))
	}(t)

	assert.NoError(t, err, "No error should happen")

	eventPodLogging := test.EventPodLogging.DeepCopy() // it is needed to prevent data race in tests
	fakeLW.Add(eventPodLogging)
	filterAllLogs := &filter.Filters{
		Sinks: []*filter.Sink{{Name: "logs"}}}
	stdoutSink, err := sink.InitStdoutSink("", filterAllLogs.GetSinkFiltersByName("logs"))
	assert.NoError(t, err, "No error should happen")

	controller := NewClusterEventController(fKubeClient, FakeListerWatcherFunc(fakeLW), []sink.ISink{stdoutSink})

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	eventPodTracing := test.EventPodTracing.DeepCopy()
	fakeLW.Add(eventPodTracing)

	//wait for the event processing some seconds
	time.Sleep(1 * time.Second)

	result, err := os.ReadFile(fname)
	assert.NoError(t, err, "No error should happen")
	assert.NotEqual(t, 0, len(result), "Stdout file should not be empty")

	expectedEventLog := strings.Builder{}
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodLogging), "No error should happen")
	assert.Equal(t, 1, strings.Count(string(result), expectedEventLog.String()), "Stdout file should contain the event from logging namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodTracing), "No error should happen")
	assert.Equal(t, 1, strings.Count(string(result), expectedEventLog.String()), "Stdout file should contain the event from tracing namespace")

	fakeLW.Delete(eventPodLogging)
	fakeLW.Delete(eventPodTracing)
}

func Test_NamespacedEventController_StdoutSink_TwoNamespaces(t *testing.T) {
	var fakeLW = fcache.NewFakeControllerSource()
	// change stdout to print events in file
	initialStdout := os.Stdout

	fname, err := test.ChangeStdoutToFile("stdout2")
	defer func(t *testing.T) {
		assert.NoError(t, test.ChangeFileToStdout(initialStdout))
	}(t)
	assert.NoError(t, err, "No error should happen")

	namespaces := []string{"logging", "tracing"}

	eventPodLogging := test.EventPodLogging.DeepCopy()
	fakeLW.Add(eventPodLogging)

	filterAllLogs := &filter.Filters{
		Sinks: []*filter.Sink{{Name: "logs"}}}
	stdoutSink, err := sink.InitStdoutSink("", filterAllLogs.GetSinkFiltersByName("logs"))
	assert.NoError(t, err, "No error should happen")

	controllers := NewNamespacedEventControllers(fKubeClient, namespaces, FakeListerWatcherFunc(fakeLW), []sink.ISink{stdoutSink})

	stop := make(chan struct{})
	defer close(stop)
	for _, c := range controllers {
		go c.Run(1, stop)
	}
	eventPodTracing := test.EventPodTracing.DeepCopy()
	fakeLW.Add(eventPodTracing)

	//wait for the event processing some seconds
	time.Sleep(1 * time.Second)

	result, err := os.ReadFile(fname)
	assert.NoError(t, err, "No error should happen")
	assert.NotEqual(t, 0, len(result), "Stdout file should not be empty")

	expectedEventLog := strings.Builder{}
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodLogging), "No error should happen")
	//fakeLW has no filtration of namespaces of objects, so there will be 2 occurrences in the file
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from logging namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodTracing), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from tracing namespace")

	fakeLW.Delete(eventPodLogging)
	fakeLW.Delete(eventPodTracing)
}

func Test_ClusterEventController_MetricsSink(t *testing.T) {
	var fakeLW = fcache.NewFakeControllerSource()
	eventPodLogging := test.EventPodLogging.DeepCopy()
	fakeLW.Add(eventPodLogging)
	defer sink.UnregisterMetrics()
	filterAllLogs := &filter.Filters{
		Sinks: []*filter.Sink{{Name: "metrics"}}}
	metricsSink, err := sink.InitMetricsSink(context.TODO(), ":9999", "", filterAllLogs.GetSinkFiltersByName("metrics"), test.StartFakeHttpServer)
	assert.NoError(t, err, "No error should happen")

	controller := NewClusterEventController(fKubeClient, FakeListerWatcherFunc(fakeLW), []sink.ISink{metricsSink})

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	eventDeploymentMonitoring := test.EventDeploymentMonitoring.DeepCopy()
	eventPodTracing := test.EventPodTracing.DeepCopy()
	eventPvcMonitoring := test.EventPvcMonitoring.DeepCopy()
	fakeLW.Add(eventDeploymentMonitoring)
	fakeLW.Add(eventPodTracing)
	fakeLW.Add(eventPvcMonitoring)

	//wait for the event processing some seconds
	time.Sleep(1 * time.Second)

	resp, err := test.FakeServer.Client().Get(test.FakeServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()
	responseBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.True(t, len(responseBody) > 0)
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"monitoring\",kind=\"Deployment\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"monitoring\",kind=\"PersistentVolumeClaim\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"logging\",kind=\"Pod\",type=\"Normal\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"tracing\",kind=\"Pod\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",event_object=\"test-pod\",kind=\"Pod\",message=\"Created or started container\",reason=\"Started\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",event_object=\"test-pod\",kind=\"Deployment\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",event_object=\"test-pvc-0\",kind=\"PersistentVolumeClaim\",message=\"storageclass not found\",reason=\"ProvisioningFailed\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",event_object=\"test-pod\",kind=\"Pod\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",kind=\"Deployment\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",kind=\"PersistentVolumeClaim\"} 1"))

	fakeLW.Delete(eventPodLogging)
	fakeLW.Delete(eventDeploymentMonitoring)
	fakeLW.Delete(eventPodTracing)
	fakeLW.Delete(eventPvcMonitoring)
}

func Test_ClusterEventController_LogsSink_MetricsSink(t *testing.T) {
	var fakeLW = fcache.NewFakeControllerSource()
	// change stdout to print events in file
	initialStdout := os.Stdout
	fname, err := test.ChangeStdoutToFile("stdout1")
	defer func(t *testing.T) {
		assert.NoError(t, test.ChangeFileToStdout(initialStdout))
	}(t)

	assert.NoError(t, err)

	eventPodLogging := test.EventPodLogging.DeepCopy()
	fakeLW.Add(eventPodLogging)
	defer sink.UnregisterMetrics()
	filterAllLogs := &filter.Filters{
		Sinks: []*filter.Sink{{Name: "metrics"}, {Name: "logs"}}}
	metricsSink, err := sink.InitMetricsSink(context.TODO(), ":9999", "", filterAllLogs.GetSinkFiltersByName("metrics"), test.StartFakeHttpServer)
	assert.NoError(t, err)
	stdoutSink, err := sink.InitStdoutSink("", filterAllLogs.GetSinkFiltersByName("logs"))
	assert.NoError(t, err)

	controller := NewClusterEventController(fKubeClient, FakeListerWatcherFunc(fakeLW), []sink.ISink{metricsSink, stdoutSink})

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	eventDeploymentMonitoring := test.EventDeploymentMonitoring.DeepCopy()
	eventPodTracing := test.EventPodTracing.DeepCopy()
	eventPvcMonitoring := test.EventPvcMonitoring.DeepCopy()
	fakeLW.Add(eventDeploymentMonitoring)
	fakeLW.Add(eventPodTracing)
	fakeLW.Add(eventPvcMonitoring)

	//wait for the event processing some seconds
	time.Sleep(1 * time.Second)

	//check stdout sink
	result, err := os.ReadFile(fname)
	assert.NoError(t, err, "No error should happen")
	assert.NotEqual(t, 0, len(result), "Stdout file should not be empty")

	expectedEventLog := strings.Builder{}
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodLogging), "No error should happen")
	assert.Equal(t, 1, strings.Count(string(result), expectedEventLog.String()), "Stdout file should contain the event from logging namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodTracing), "No error should happen")
	assert.Equal(t, 1, strings.Count(string(result), expectedEventLog.String()), "Stdout file should contain the event from tracing namespace")

	//check metrics sink
	resp, err := test.FakeServer.Client().Get(test.FakeServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()
	responseBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.True(t, len(responseBody) > 0)
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"monitoring\",kind=\"Deployment\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"monitoring\",kind=\"PersistentVolumeClaim\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"logging\",kind=\"Pod\",type=\"Normal\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"tracing\",kind=\"Pod\",type=\"Warning\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",event_object=\"test-pod\",kind=\"Pod\",message=\"Created or started container\",reason=\"Started\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",event_object=\"test-pod\",kind=\"Deployment\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",event_object=\"test-pvc-0\",kind=\"PersistentVolumeClaim\",message=\"storageclass not found\",reason=\"ProvisioningFailed\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",event_object=\"test-pod\",kind=\"Pod\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",kind=\"Deployment\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",kind=\"PersistentVolumeClaim\"} 1"))

	fakeLW.Delete(eventPodLogging)
	fakeLW.Delete(eventDeploymentMonitoring)
	fakeLW.Delete(eventPodTracing)
	fakeLW.Delete(eventPvcMonitoring)
}
