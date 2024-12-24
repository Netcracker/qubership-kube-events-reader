package aggregation

import "regexp"

var rsAggregationRegexps = map[int]*regexp.Regexp{}
var rsAggregationLabelValues = map[int]string{}

var rsAggregations = map[string]string{
	"Deleted pod: .*":                    "Deleted pod",
	"Created pod: .*":                    "Created pod",
	"Error deleting: .*":                 "Error deleting",
	"Error creating: pods .*forbidden.*": "Error creating: forbidden",
	"Error creating: .*":                 "Error creating",
}

var depAggregationRegexps = map[int]*regexp.Regexp{}
var depAggregationLabelValues = map[int]string{}

var depAggregations = map[string]string{
	"Scaled down replica set .* to \\d+.*":                                      "Scaled down replica set",
	"Scaled up replica set .* to \\d+.*":                                        "Scaled up replica set",
	"The rollback revision contains the same template as current deployment .*": "The rollback revision contains the same template as current deployment",
	"Rolled back deployment .* to revision .*":                                  "Rolled back deployment to previous revision",
	"Failed to create new replica set .*":                                       "Failed to create new replica set",
}

var depConfigAggregationRegexps = map[int]*regexp.Regexp{}
var depConfigAggregationLabelValues = map[int]string{}
var depConfigAggregations = map[string]string{
	"Rollout for .* cancelled":             "Rollout cancelled",
	"Created new replication controller.*": "Created new replication controller",
	"Cancelled deployment.*":               "Cancelled deployment",
	"Deployment of version \\d+ awaiting cancellation of older running deployments": "Deployment awaiting cancellation of older running deployments",
}
