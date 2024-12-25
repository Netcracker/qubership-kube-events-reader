package aggregation

import "regexp"

var hpaAggregationRegexps = map[int]*regexp.Regexp{}
var hpaAggregationLabelValues = map[int]string{}
var hpaAggregations = map[string]string{
	".*couldn't convert selector into a corresponding internal selector object.*": "Couldn't convert selector",
	".*pods by selector .* are controlled by multiple HPAs.*":                     "Pods are controlled by multiple HPAs",
	"New size: \\d+; reason: .*":                                                  "New size",
}
