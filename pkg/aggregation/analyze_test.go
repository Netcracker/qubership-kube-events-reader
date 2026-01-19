package aggregation

import (
	"regexp"
	"testing"
)

func TestGetCommonMessage(t *testing.T) {
	// Initialize aggregations first
	InitAggregations()

	tests := []struct {
		name     string
		kind     string
		reason   string
		message  string
		expected string
	}{
		{"pod with matching message", "pod", "Failed", "Successfully pulled image nginx:latest", "Successfully pulled image"},
		{"pod with non-matching", "pod", "Failed", "Some unknown message", "Some unknown message"},
		{"unknown kind", "unknown", "Failed", "Some message", "Some message"},
		{"owner ref invalid namespace", "pod", "OwnerRefInvalidNamespace", "ownerRef test does not exist in namespace test", "ownerRef does not exist in namespace"},
		{"forbidden update", "unknown", "UpdateError", "is forbidden: User test cannot update resource test in API group", "Forbidden: User cannot update resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommonMessage(tt.kind, tt.reason, tt.message)
			if result != tt.expected {
				t.Errorf("GetCommonMessage(%q, %q, %q) = %q, want %q", tt.kind, tt.reason, tt.message, result, tt.expected)
			}
		})
	}
}

func TestGetCommonMessageForEvent(t *testing.T) {
	// Create test regexps and labels
	testRegexps := make(map[int]*regexp.Regexp)
	testLabels := make(map[int]string)
	testAggregations := map[string]string{
		"Successfully pulled image .*": "Successfully pulled image",
	}

	it := 0
	for expression, value := range testAggregations {
		testRegexps[it] = regexp.MustCompile(expression)
		testLabels[it] = value
		it++
	}

	tests := []struct {
		name     string
		reason   string
		message  string
		expected string
	}{
		{"matching pattern", "SomeReason", "Successfully pulled image nginx:latest", "Successfully pulled image"},
		{"non-matching pattern", "SomeReason", "Some random message", "Some random message"},
		{"owner ref invalid", "OwnerRefInvalidNamespace", "Some message", "ownerRef does not exist in namespace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommonMessageForEvent(tt.reason, tt.message, testRegexps, testLabels)
			if result != tt.expected {
				t.Errorf("getCommonMessageForEvent(%q, %q) = %q, want %q", tt.reason, tt.message, result, tt.expected)
			}
		})
	}
}

func TestGetCommonMessageForHPA(t *testing.T) {
	tests := []struct {
		name     string
		reason   string
		message  string
		expected string
	}{
		{"FailedGetScale", "FailedGetScale", "some message", "FailedGetScale"},
		{"FailedComputeMetricsReplicas", "FailedComputeMetricsReplicas", "some message", "FailedComputeMetricsReplicas"},
		{"other reason", "OtherReason", "some message", "some message"},
	}

	// Empty maps for this test
	emptyRegexps := make(map[int]*regexp.Regexp)
	emptyLabels := make(map[int]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommonMessageForHPA(tt.reason, tt.message, emptyRegexps, emptyLabels)
			if result != tt.expected {
				t.Errorf("getCommonMessageForHPA(%q, %q) = %q, want %q", tt.reason, tt.message, result, tt.expected)
			}
		})
	}
}

func TestGetAllKindsMessageByReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   string
		message  string
		expected string
	}{
		{"OwnerRefInvalidNamespace with regexp match", "OwnerRefInvalidNamespace", "ownerRef test does not exist in namespace test", "ownerRef does not exist in namespace"},
		{"OwnerRefInvalidNamespace no match", "OwnerRefInvalidNamespace", "some other message", "some other message"},
		{"UpdateError with regexp match", "UpdateError", "is forbidden: User test cannot update resource test in API group", "Forbidden: User cannot update resource"},
		{"UpdateError no match", "UpdateError", "some other message", "some other message"},
		{"other reason", "OtherReason", "some message", "some message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAllKindsMessageByReason(tt.reason, tt.message)
			if result != tt.expected {
				t.Errorf("getAllKindsMessageByReason(%q, %q) = %q, want %q", tt.reason, tt.message, result, tt.expected)
			}
		})
	}
}
