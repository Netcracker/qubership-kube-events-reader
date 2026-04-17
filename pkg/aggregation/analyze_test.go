package aggregation

import (
	"regexp"
	"testing"
)

func TestInitAggregationsForKind(t *testing.T) {
	aggregations := map[string]string{
		"foo .*": "foo",
		"bar .*": "bar",
	}
	regexps := map[int]*regexp.Regexp{}
	labelValues := map[int]string{}

	initAggregationsForKind(aggregations, regexps, labelValues)

	if len(aggregations) != 0 {
		t.Fatalf("expected source aggregations to be cleared, got %d entries", len(aggregations))
	}
	if len(regexps) != 2 {
		t.Fatalf("expected 2 compiled regexps, got %d", len(regexps))
	}
	if len(labelValues) != 2 {
		t.Fatalf("expected 2 label values, got %d", len(labelValues))
	}

	matched := 0
	for index, expression := range regexps {
		if expression.MatchString("foo event") && labelValues[index] == "foo" {
			matched++
		}
		if expression.MatchString("bar event") && labelValues[index] == "bar" {
			matched++
		}
	}
	if matched != 2 {
		t.Fatalf("expected both label mappings to be preserved, got %d successful matches", matched)
	}
}

func TestGetCommonMessageForEvent(t *testing.T) {
	regexps := map[int]*regexp.Regexp{
		0: regexp.MustCompile("matched .*"),
	}
	labelValues := map[int]string{
		0: "normalized message",
	}

	if got := getCommonMessageForEvent("AnyReason", "matched event", regexps, labelValues); got != "normalized message" {
		t.Fatalf("expected regexp-based message, got %q", got)
	}
	if got := getCommonMessageForEvent("OwnerRefInvalidNamespace", "original", regexps, labelValues); got != "ownerRef does not exist in namespace" {
		t.Fatalf("expected ownerRef fallback, got %q", got)
	}
	if got := getCommonMessageForEvent("OtherReason", "original", regexps, labelValues); got != "original" {
		t.Fatalf("expected original message to be preserved, got %q", got)
	}
}

func TestGetCommonMessageForHPA(t *testing.T) {
	regexps := map[int]*regexp.Regexp{
		0: regexp.MustCompile("scale .*"),
	}
	labelValues := map[int]string{
		0: "scaled",
	}

	if got := getCommonMessageForHPA("FailedGetScale", "ignored", regexps, labelValues); got != "FailedGetScale" {
		t.Fatalf("expected special FailedGetScale handling, got %q", got)
	}
	if got := getCommonMessageForHPA("FailedComputeMetricsReplicas", "ignored", regexps, labelValues); got != "FailedComputeMetricsReplicas" {
		t.Fatalf("expected special FailedComputeMetricsReplicas handling, got %q", got)
	}
	if got := getCommonMessageForHPA("OtherReason", "scale event", regexps, labelValues); got != "scaled" {
		t.Fatalf("expected HPA to reuse generic event normalization, got %q", got)
	}
}

func TestGetAllKindsMessageByReason(t *testing.T) {
	ownerRefMessage := "ownerRef test-owner does not exist in namespace monitoring"
	if got := getAllKindsMessageByReason("OwnerRefInvalidNamespace", ownerRefMessage); got != "ownerRef does not exist in namespace" {
		t.Fatalf("expected ownerRef normalization, got %q", got)
	}

	forbiddenMessage := `event is forbidden: User "system:serviceaccount:test" cannot update resource "events" in API group`
	if got := getAllKindsMessageByReason("UpdateError", forbiddenMessage); got != "Forbidden: User cannot update resource" {
		t.Fatalf("expected forbidden update normalization, got %q", got)
	}

	if got := getAllKindsMessageByReason("OtherReason", "original"); got != "original" {
		t.Fatalf("expected original message for unmatched reason, got %q", got)
	}
}

func TestGetCommonMessageRoutesByKind(t *testing.T) {
	originalPodRegexps := podAggregationRegexps
	originalPodLabels := podAggregationLabelValues
	originalIssuerRegexps := issuerAggregationRegexps
	originalIssuerLabels := issuerAggregationLabelValues
	defer func() {
		podAggregationRegexps = originalPodRegexps
		podAggregationLabelValues = originalPodLabels
		issuerAggregationRegexps = originalIssuerRegexps
		issuerAggregationLabelValues = originalIssuerLabels
	}()

	podAggregationRegexps = map[int]*regexp.Regexp{
		0: regexp.MustCompile("pod event .*"),
	}
	podAggregationLabelValues = map[int]string{
		0: "pod-normalized",
	}
	issuerAggregationRegexps = map[int]*regexp.Regexp{
		0: regexp.MustCompile("issuer event .*"),
	}
	issuerAggregationLabelValues = map[int]string{
		0: "issuer-normalized",
	}

	if got := GetCommonMessage("PoD", "AnyReason", "pod event happened"); got != "pod-normalized" {
		t.Fatalf("expected pod route to use pod aggregation tables, got %q", got)
	}
	if got := GetCommonMessage("issuer", "ErrGetKeyPair", "ignored"); got != "Error getting keypair for CA issuer" {
		t.Fatalf("expected issuer special case, got %q", got)
	}
	if got := GetCommonMessage("ClusterIssuer", "OtherReason", "issuer event happened"); got != "issuer-normalized" {
		t.Fatalf("expected cluster issuer route to use cert-manager tables, got %q", got)
	}
	if got := GetCommonMessage("UnknownKind", "UpdateError", `x is forbidden: User "u" cannot update resource "r" in API group`); got != "Forbidden: User cannot update resource" {
		t.Fatalf("expected unknown kinds to use generic fallback rules, got %q", got)
	}
}
