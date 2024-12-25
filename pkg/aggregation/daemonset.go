package aggregation

import "regexp"

var dsAggregationRegexps = map[int]*regexp.Regexp{}
var dsAggregationLabelValues = map[int]string{}

var dsAggregations = map[string]string{
	"Created pod.*":                       "Created pod",
	"Error creating: pods .* forbidden.*": "Error creating pods - forbidden",
	"Found failed daemon pod .* on node .*, will try to kill it":      "Found failed daemon pod on node, will try to kill it",
	"Found succeeded daemon pod .* on node .*, will try to delete it": "Found succeeded daemon pod on node, will try to delete it",
}
