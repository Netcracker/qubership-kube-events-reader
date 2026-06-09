package sink

import (
	"fmt"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/format"
	corev1 "k8s.io/api/core/v1"
)

type StdoutSink struct {
	*Sink
}

func InitStdoutSink(printFormat string, filters *filter.Sink) (*StdoutSink, error) {
	err := format.SetFormat(printFormat)
	if err != nil {
		return nil, err
	}
	sink := initializeSinkWithFilters(filters)
	return &StdoutSink{Sink: sink}, nil
}

func (ss *StdoutSink) Release(eventObj *corev1.Event) error {
	if !ss.IsEventAllowed(eventObj) {
		return nil
	}
	fmt.Println(format.FormatEvent(eventObj))
	return nil
}
