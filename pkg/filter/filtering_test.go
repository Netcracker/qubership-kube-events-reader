package filter

import (
	"flag"
	"reflect"
	"testing"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/aggregation"
	"github.com/stretchr/testify/assert"
)

func Test_ParseFiltersConfiguration(t *testing.T) {
	filters, err := ParseFiltersConfiguration("../test/filtering_config_valid.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, filters)
	expectedFilters := &Filters{
		Sinks: []*Sink{
			{
				Name: "metrics",
				Match: []EventMatch{
					{Type: "Warning", Kind: "Pod|Deployment"},
				},
				Exclude: []EventMatch{
					{Type: "Normal", Message: ".*image.*"},
					{Reason: "Completed|Pulled|Started"},
				},
			},
			{
				Name: "logs",
				Match: []EventMatch{
					{Type: "Warning"},
				},
				Exclude: []EventMatch{
					{ReportingController: "nginx-ingress-controller"},
				},
			},
		},
		MessageAggregation: aggregation.Config{
			MaxMessageLength: 128,
			Rules: []aggregation.Rule{
				{Kind: "^SparkApplication$", Message: "^failed to submit SparkApplication ", Value: "Failed to submit SparkApplication"},
			},
		},
	}
	assert.True(t, reflect.DeepEqual(filters, expectedFilters))
}

func Test_GetSinkFiltersByName(t *testing.T) {
	sink := &Sink{
		Name:    "test",
		Match:   make([]EventMatch, 0),
		Exclude: []EventMatch{{Kind: "GrafanaFolder"}},
	}
	filters := &Filters{Sinks: []*Sink{sink}}

	sinkT := filters.GetSinkFiltersByName("test")
	assert.NotNil(t, sinkT)
	assert.Equal(t, sink, sinkT)
	assert.Nil(t, filters.GetSinkFiltersByName("test1"))
}

func Test_ValidateFileSize(t *testing.T) {
	const maxFileSize = 5 * 1024 * 1024 // 5 MB
	filterFile := flag.String("filtersPath", "../test/filtering_config_valid.yaml", "Absolute path to file with filter events configuration")
	err := ValidateFileSize(*filterFile, maxFileSize)
	assert.NoError(t, err)
}
