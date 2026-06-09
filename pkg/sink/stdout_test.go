package sink

import (
	"os"
	"strings"
	"testing"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/format"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestStdoutSink_InitMetricsSink_Release_WithoutFilters(t *testing.T) {
	var filtersSink = filter.Sink{}
	testSink, err := InitStdoutSink("", &filtersSink)
	assert.NoError(t, err)
	assert.NotNil(t, testSink)
	assert.NotNil(t, testSink.Sink)
	assert.Equal(t, 0, len(testSink.Exclude))
	assert.Equal(t, 0, len(testSink.Match))

	// change stdout to print events in file
	initialStdout := os.Stdout
	fname, err := test.ChangeStdoutToFile("stdout3")
	defer func(t *testing.T) {
		assert.NoError(t, test.ChangeFileToStdout(initialStdout))
	}(t)
	assert.NoError(t, err, "No error should happen")

	for _, event := range test.TestEventsSlice {
		assert.NoError(t, testSink.Release(event))
	}

	result, err := os.ReadFile(fname)
	assert.NoError(t, err, "No error should happen")
	assert.NotEqual(t, 0, len(result), "Stdout file should not be empty")

	expectedEventLog := strings.Builder{}
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodLogging), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from logging namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodTracing), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from tracing namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventDeploymentMonitoring), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from monitoring namespace with Deployment kind of involved object")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPvcMonitoring), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from monitoring namespace with PVC kind of involved object")

}

func TestStdoutSink_InitMetricsSink_Release_WithFilters(t *testing.T) {
	testSink, err := InitStdoutSink("", &filtersSinkMatchAndExclude)
	assert.NoError(t, err)
	assert.NotNil(t, testSink)
	assert.NotNil(t, testSink.Sink)
	assert.Equal(t, 1, len(testSink.Exclude))
	assert.Equal(t, 2, len(testSink.Match))

	// change stdout to print events in file
	initialStdout := os.Stdout
	fname, err := test.ChangeStdoutToFile("stdout4")
	defer func(t *testing.T) {
		assert.NoError(t, test.ChangeFileToStdout(initialStdout))
	}(t)
	assert.NoError(t, err, "No error should happen")

	for _, event := range test.TestEventsSlice {
		assert.NoError(t, testSink.Release(event))
	}

	result, err := os.ReadFile(fname)
	assert.NoError(t, err, "No error should happen")
	assert.NotEqual(t, 0, len(result), "Stdout file should not be empty")

	expectedEventLog := strings.Builder{}
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodLogging), "No error should happen")
	assert.False(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event with type normal")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPodTracing), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from tracing namespace")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventDeploymentMonitoring), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from monitoring namespace with Deployment kind of involved object")

	expectedEventLog.Reset()
	assert.NoError(t, format.FormatTemplate.Execute(&expectedEventLog, test.EventPvcMonitoring), "No error should happen")
	assert.True(t, strings.Contains(string(result), expectedEventLog.String()), "Stdout file should contain the event from monitoring namespace with PVC kind of involved object")

}
