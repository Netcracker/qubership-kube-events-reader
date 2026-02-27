package sink

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMetricsSink_InitMetricsSink_Release_WithoutFilters(t *testing.T) {
	var filtersSink = filter.Sink{}
	testSink, err := InitMetricsSink(context.Background(), ":9999", "", &filtersSink, test.StartFakeHttpServer)
	defer test.FakeServer.Close()
	defer UnregisterMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, testSink)
	assert.NotNil(t, testSink.Sink)
	assert.Equal(t, 0, len(testSink.Exclude))
	assert.Equal(t, 0, len(testSink.Match))
	for _, event := range test.TestEventsSlice {
		assert.NoError(t, testSink.Release(event))
	}

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
}

func TestPrometheusMetricsSink_InitMetricsSink_Release_WithFilters(t *testing.T) {
	testSink, err := InitMetricsSink(context.Background(), ":9999", "", &filtersSinkMatchAndExclude, test.StartFakeHttpServer)
	defer test.FakeServer.Close()
	defer UnregisterMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, testSink)
	assert.NotNil(t, testSink.Sink)
	assert.Equal(t, 1, len(testSink.Exclude))
	assert.Equal(t, 2, len(testSink.Match))
	for _, event := range test.TestEventsSlice {
		assert.NoError(t, testSink.Release(event))
	}

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
	assert.False(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"logging\",kind=\"Pod\",type=\"Normal\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total{event_namespace=\"tracing\",kind=\"Pod\",type=\"Warning\"} 1"))
	assert.False(t, strings.Contains(string(responseBody), "kube_events_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",event_object=\"test-pod\",kind=\"Pod\",message=\"Created or started container\",reason=\"Started\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",event_object=\"test-pod\",kind=\"Deployment\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",event_object=\"test-pvc-0\",kind=\"PersistentVolumeClaim\",message=\"storageclass not found\",reason=\"ProvisioningFailed\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",event_object=\"test-pod\",kind=\"Pod\",message=\"Back-off restarting failed container\",reason=\"BackOff\"} 1"))
	assert.False(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_normal_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"logging\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"deployment-controller\",controller_instance=\"10.10.10.10\",event_namespace=\"monitoring\",kind=\"Deployment\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"kubelet\",controller_instance=\"10.10.10.10\",event_namespace=\"tracing\",kind=\"Pod\"} 1"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total{controller=\"persistentvolume-controller\",controller_instance=\"\",event_namespace=\"monitoring\",kind=\"PersistentVolumeClaim\"} 1"))
}
