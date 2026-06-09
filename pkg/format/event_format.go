package format

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"strings"
	"text/template"
)

var FormatTemplate *template.Template

var defaultFormat = "{\"time\":\"{{.LastTimestamp.Format \"2006-01-02T15:04:05.999\"}}\",\"involvedObjectKind\":\"{{.InvolvedObject.Kind}}\",\"involvedObjectNamespace\":\"{{.InvolvedObject.Namespace}}\",\"involvedObjectName\":\"{{.InvolvedObject.Name}}\",\"involvedObjectUid\":\"{{.InvolvedObject.UID}}\",\"involvedObjectApiVersion\":\"{{.InvolvedObject.APIVersion}}\",\"involvedObjectResourceVersion\":\"{{.InvolvedObject.ResourceVersion}}\",\"reason\":\"{{.Reason}}\",\"type\":\"{{.Type}}\",\"message\":\"{{js .Message}}\",\"kind\":\"KubernetesEvent\"}"

// SetFormat initializes text template to print logs of events
func SetFormat(format string) error {
	if len(format) == 0 || len(strings.TrimSpace(format)) == 0 {
		slog.Warn("Template format is not set. Default is used.")
		t, err := template.New("format").Parse(defaultFormat)
		FormatTemplate = t
		if err != nil {
			slog.Error("Failed when creating default template", "error", err)
			return err
		}
		return nil
	}
	t, err := template.New("format").Parse(format)
	if err != nil {
		return err
	}
	FormatTemplate = t
	return nil
}

// FormatEvent returns formatted string of given Event using predefined template
func FormatEvent(event *corev1.Event) (formatted string) {

	writer := strings.Builder{}

	if event.LastTimestamp.IsZero() {
		slog.Debug("Event lastTimestamp is zero. Using current timestamp", "Event", event)
		event.LastTimestamp = metav1.Now()
	}

	if err := FormatTemplate.Execute(&writer, event); err != nil {
		slog.Error("Could not execute template for Event", "error", err)
		return
	}
	formatted = writer.String()
	return
}
