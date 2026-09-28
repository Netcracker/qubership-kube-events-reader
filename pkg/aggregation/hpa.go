package aggregation

import "regexp"

var hpaAggregationRegexps = map[int]*regexp.Regexp{}
var hpaAggregationLabelValues = map[int]string{}
var hpaAggregations = map[string]string{
	".*couldn't convert selector into a corresponding internal selector object.*": "Couldn't convert selector",
	".*pods by selector .* are controlled by multiple HPAs.*":                     "Pods are controlled by multiple HPAs",
	"New size: \\d+; reason: .*":                                                  "New size",
	"failed to get (\\w+) utilization.*":                                          "Failed to get ${1} utilization",
	"failed to get (\\w+) usage.*":                                                "Failed to get ${1} usage",
	"unable to get metric ([^:\\s]+).*":                                           "Unable to get metric ${1}",
	"unable to get external metric [^/]*/([^/]+)/.*":                              "Unable to get external metric ${1}",
}
