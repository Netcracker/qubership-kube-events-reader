package aggregation

import "regexp"

var grafanaAggregationRegexps = map[int]*regexp.Regexp{}
var grafanaAggregationLabelValues = map[int]string{}
var grafanaAggregations = map[string]string{
	"dashboard .* successfully submitted": "dashboard successfully submitted",
}
