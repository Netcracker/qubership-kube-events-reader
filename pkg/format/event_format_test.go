package format

import (
	"strings"
	"testing"
	"text/template"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/stretchr/testify/assert"
)

func Test_EventFormat_Default(t *testing.T) {

	err := SetFormat("")
	assert.NoError(t, err, "No error should happen")

	templ, err := template.New("test").Parse(defaultFormat)
	assert.NoError(t, err, "No error should happen")

	expectedFormattedEvent := strings.Builder{}
	err = templ.Execute(&expectedFormattedEvent, test.EventPodLogging)
	assert.NoError(t, err, "No error should happen")

	formattedEvent := FormatEvent(test.EventPodLogging)
	assert.Equal(t, 0, strings.Compare(expectedFormattedEvent.String(), formattedEvent), "Formatted event should be printed using default template")
}

var eventFormatTest = "time={{.LastTimestamp.Format \"2006-01-02T15:04:05Z\"}} involvedObject.kind={{.InvolvedObject.Kind}} involvedObject.namespace={{.InvolvedObject.Namespace}} involvedObject.name={{.InvolvedObject.Name}} involvedObject.uid={{.InvolvedObject.UID}} involvedObject.apiVersion={{.InvolvedObject.APIVersion}} involvedObject.resourceVersion={{.InvolvedObject.ResourceVersion}} reason={{.Reason}} message=\"{{js .Message}}\" firstTimestamp={{.FirstTimestamp.Format \"2006-01-02T15:04:05Z\"}} lastTimestamp={{.LastTimestamp.Format \"2006-01-02T15:04:05Z\"}} count={{.Count}} type={{.Type}} eventTime={{ if not .EventTime.IsZero }}{{.EventTime.Format \"2006-01-02T15:04:05Z\"}}{{end}} kind=EventTest"

func Test_EventFormat_Custom(t *testing.T) {

	err := SetFormat(eventFormatTest)
	assert.NoError(t, err, "No error should happen")

	templ, err := template.New("test").Parse(eventFormatTest)
	assert.NoError(t, err, "No error should happen")

	expectedFormattedEvent := strings.Builder{}
	err = templ.Execute(&expectedFormattedEvent, test.EventPodLogging)
	assert.NoError(t, err, "No error should happen")

	formattedEvent := FormatEvent(test.EventPodLogging)
	assert.Equal(t, 0, strings.Compare(expectedFormattedEvent.String(), formattedEvent), "Formatted event should be printed using default template")
}
