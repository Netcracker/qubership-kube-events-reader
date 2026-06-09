package sink

import (
	"regexp"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	corev1 "k8s.io/api/core/v1"
)

type Sink struct {
	Match   []*Rule
	Exclude []*Rule
}

type Rule struct {
	Type                *regexp.Regexp
	Kind                *regexp.Regexp
	Reason              *regexp.Regexp
	Namespace           *regexp.Regexp
	ReportingController *regexp.Regexp
	ReportingInstance   *regexp.Regexp
	Message             *regexp.Regexp
}

type ISink interface {
	Release(*corev1.Event) error
	IsEventAllowed(*corev1.Event) bool
}

func (s *Sink) IsEventAllowed(eventObj *corev1.Event) bool {
	for _, e := range s.Exclude {
		if e.isEventToBeExcluded(eventObj) {
			return false
		}
	}
	match := true
	for _, e := range s.Match {
		match = e.isEventMatched(eventObj)
		if match {
			return true
		}
	}
	return match
}

func (rule *Rule) isEventToBeExcluded(eventObj *corev1.Event) bool {
	exclude := false
	if rule.Type != nil {
		exclude = rule.Type.MatchString(eventObj.Type)
	}
	if !exclude && rule.Kind != nil {
		exclude = rule.Kind.MatchString(eventObj.InvolvedObject.Kind)
	}
	if !exclude && rule.Namespace != nil {
		exclude = rule.Namespace.MatchString(eventObj.InvolvedObject.Namespace)
	}
	if !exclude && rule.Reason != nil {
		exclude = rule.Reason.MatchString(eventObj.Reason)
	}
	if !exclude && rule.Message != nil {
		exclude = rule.Message.MatchString(eventObj.Message)
	}
	if !exclude && rule.ReportingController != nil {
		exclude = rule.ReportingController.MatchString(eventObj.ReportingController)
	}
	if !exclude && rule.ReportingInstance != nil {
		exclude = rule.ReportingInstance.MatchString(eventObj.ReportingInstance)
	}
	return exclude
}

func (rule *Rule) isEventMatched(eventObj *corev1.Event) bool {
	match := true
	if rule.Type != nil {
		match = rule.Type.MatchString(eventObj.Type)
	}
	if match && rule.Kind != nil {
		match = rule.Kind.MatchString(eventObj.InvolvedObject.Kind)
	}
	if match && rule.Namespace != nil {
		match = rule.Namespace.MatchString(eventObj.InvolvedObject.Namespace)
	}
	if match && rule.Reason != nil {
		match = rule.Reason.MatchString(eventObj.Reason)
	}
	if match && rule.Message != nil {
		match = rule.Message.MatchString(eventObj.Message)
	}
	if match && rule.ReportingController != nil {
		match = rule.ReportingController.MatchString(eventObj.ReportingController)
	}
	if match && rule.ReportingInstance != nil {
		match = rule.ReportingInstance.MatchString(eventObj.ReportingInstance)
	}
	return match
}

func initializeSinkWithFilters(filters *filter.Sink) *Sink {
	var sink Sink
	if filters == nil {
		return &sink
	}
	if filters.Match != nil {
		sink.Match = make([]*Rule, len(filters.Match))
	}
	if filters.Exclude != nil {
		sink.Exclude = make([]*Rule, len(filters.Exclude))
	}

	for i, s := range filters.Match {
		rule := createRule(&s)
		sink.Match[i] = rule
	}
	for i, s := range filters.Exclude {
		rule := createRule(&s)
		sink.Exclude[i] = rule
	}
	return &sink
}

func createRule(eventMatch *filter.EventMatch) *Rule {
	rule := Rule{}
	if len(eventMatch.Type) > 0 {
		rule.Type = regexp.MustCompile(eventMatch.Type)
	}
	if len(eventMatch.Kind) > 0 {
		rule.Kind = regexp.MustCompile(eventMatch.Kind)
	}
	if len(eventMatch.Namespace) > 0 {
		rule.Namespace = regexp.MustCompile(eventMatch.Namespace)
	}
	if len(eventMatch.Reason) > 0 {
		rule.Reason = regexp.MustCompile(eventMatch.Reason)
	}
	if len(eventMatch.Message) > 0 {
		rule.Message = regexp.MustCompile(eventMatch.Message)
	}
	if len(eventMatch.ReportingController) > 0 {
		rule.ReportingController = regexp.MustCompile(eventMatch.ReportingController)
	}
	if len(eventMatch.ReportingInstance) > 0 {
		rule.ReportingInstance = regexp.MustCompile(eventMatch.ReportingInstance)
	}
	return &rule
}
