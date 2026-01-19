package test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestEventsSlice = []*corev1.Event{EventPodLogging, EventPodTracing, EventDeploymentMonitoring, EventPvcMonitoring}

var LastTs = metav1.Time{Time: time.Now()}
var EventPodLogging = &corev1.Event{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-event-28495390-fvlw8.17ba285fedd400ee",
		Namespace: "logging",
	},
	InvolvedObject: corev1.ObjectReference{
		Kind:            "Pod",
		Namespace:       "logging",
		Name:            "test-pod",
		UID:             "04e98ff7-7471-451f-a9cf-4bcad4a1bd41",
		APIVersion:      "v1",
		ResourceVersion: "78496715",
	},
	Message:             "Started container test",
	FirstTimestamp:      LastTs,
	LastTimestamp:       LastTs,
	Type:                "Normal",
	Reason:              "Started",
	ReportingController: "kubelet",
	ReportingInstance:   "10.10.10.10",
	Source: corev1.EventSource{
		Component: "k8s-fake-test",
	},
	Count: 5,
}

var EventPodTracing = &corev1.Event{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-event-9549b7b5d-b7vrs.178f3c5ceebf55de",
		Namespace: "tracing",
	},
	InvolvedObject: corev1.ObjectReference{
		Kind:            "Pod",
		Namespace:       "tracing",
		Name:            "test-pod",
		UID:             "4d665856-6f0a-46dc-ac63-7c86d5a21a1c",
		APIVersion:      "v1",
		ResourceVersion: "80427479",
	},
	Message:             "Back-off restarting failed container",
	FirstTimestamp:      LastTs,
	LastTimestamp:       LastTs,
	Type:                "Warning",
	Reason:              "BackOff",
	ReportingController: "kubelet",
	ReportingInstance:   "10.10.10.10",
	Source: corev1.EventSource{
		Component: "k8s-fake-test",
	},
	Count: 3,
}

var EventDeploymentMonitoring = &corev1.Event{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-event-deployment-566cbb5bd4.17e2b744ee79ba03",
		Namespace: "monitoring",
	},
	InvolvedObject: corev1.ObjectReference{
		Kind:            "Deployment",
		Namespace:       "monitoring",
		Name:            "test-pod",
		UID:             "4d665856-6f0a-46dc-ac63-7c86d5a21a1c",
		APIVersion:      "v1",
		ResourceVersion: "80427479",
	},
	Message:             "Back-off restarting failed container",
	FirstTimestamp:      LastTs,
	LastTimestamp:       LastTs,
	Type:                "Warning",
	Reason:              "BackOff",
	ReportingController: "deployment-controller",
	ReportingInstance:   "10.10.10.10",
	Source: corev1.EventSource{
		Component: "k8s-fake-test",
	},
	Count: 3,
}

var EventPvcMonitoring = &corev1.Event{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-event-pvc-pvc-0.17deefba63888faa",
		Namespace: "monitoring",
	},
	InvolvedObject: corev1.ObjectReference{
		Kind:            "PersistentVolumeClaim",
		Namespace:       "monitoring",
		Name:            "test-pvc-0",
		UID:             "a850b070-cd63-442e-915f-b6f1cc4bcdae",
		APIVersion:      "v1",
		ResourceVersion: "888879630",
	},
	Message:        "storageclass.storage.k8s.io \"csi-cinder-sc-delete\" not found",
	FirstTimestamp: LastTs,
	LastTimestamp:  LastTs,
	Type:           "Warning",
	Reason:         "ProvisioningFailed",
	Source: corev1.EventSource{
		Component: "persistentvolume-controller",
	},
	ReportingController: "persistentvolume-controller",
	Count:               98494,
}

func ChangeStdoutToFile(fileName string) (string, error) {
	fname := filepath.Join(os.TempDir(), fileName)
	temp, err := os.Create(fname)
	if err != nil {
		return "", err
	}
	os.Stdout = temp
	return fname, nil
}

func ChangeFileToStdout(initialStdout *os.File) error {
	fileStdout := os.Stdout
	os.Stdout = initialStdout
	if err := fileStdout.Close(); err != nil {
		return err
	}
	err := os.Remove(fileStdout.Name())
	return err
}

var FakeServer = &httptest.Server{}

func StartFakeHttpServer(ctx context.Context, port string) {
	FakeServer = httptest.NewServer(promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
}
