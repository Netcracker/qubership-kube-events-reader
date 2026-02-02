package sink

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
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
	assert.True(t, strings.Contains(string(responseBody), "kube_events_total"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_normal_total"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_warning_total"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_normal_total"))
	assert.True(t, strings.Contains(string(responseBody), "kube_events_reporting_controller_warning_total"))
}

func TestPrometheusMetricsSink_InitMetricsSink_Release_WithFilters(t *testing.T) {
	registry := prometheus.NewRegistry()
	testSink, err := InitMetricsSinkWithRegistry(context.Background(), "9999", "/metrics", &filtersSinkMatchAndExclude, nil, registry)
	defer UnregisterMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, testSink)
	assert.NotNil(t, testSink.Sink)
	assert.Equal(t, 1, len(testSink.Exclude))
	assert.Equal(t, 2, len(testSink.Match))
	for _, event := range test.TestEventsSlice {
		assert.NoError(t, testSink.Release(event))
	}

	// Gather metrics from the registry
	gatherers := prometheus.Gatherers{registry}
	metrics, err := gatherers.Gather()
	assert.NoError(t, err)

	// Convert to text format
	var buf bytes.Buffer
	for _, mf := range metrics {
		_, err := expfmt.MetricFamilyToText(&buf, mf)
		assert.NoError(t, err)
	}
	responseBody := buf.String()

	assert.True(t, len(responseBody) > 0)
	assert.True(t, strings.Contains(responseBody, "kube_events_total"))
	assert.True(t, strings.Contains(responseBody, "kube_events_warning_total"))
	assert.True(t, strings.Contains(responseBody, "kube_events_reporting_controller_warning_total"))
	// With filters, the normal events should be filtered out
	assert.False(t, strings.Contains(responseBody, "kube_events_normal_total"))
	assert.False(t, strings.Contains(responseBody, "kube_events_reporting_controller_normal_total"))
}
