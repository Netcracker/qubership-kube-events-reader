package aggregation

import "regexp"

var jobAggregationRegexps = map[int]*regexp.Regexp{}
var jobAggregationLabelValues = map[int]string{}

var jobAggregations = map[string]string{
	"Created pod: .*":               "Created pod",
	"Deleted pod: .*":               "Deleted pod",
	"Error deleting: .*":            "Error deleting",
	"Error creating: .*":            "Error creating",
	"Error creating: .*forbidden.*": "Error creating: forbidden",
}

var cronjobAggregationRegexps = map[int]*regexp.Regexp{}
var cronjobAggregationLabelValues = map[int]string{}

var cronjobAggregations = map[string]string{
	"Created job.*":                      "Created job",
	"Deleted job.*":                      "Deleted job",
	"unparseable schedule for cronjob.*": "Unparseable schedule for cronjob",
	"Saw a job that the controller did not create or forgot.*": "Saw a job that the controller did not create or forgot",
	"Saw completed job.*":                    "Saw completed job",
	"Active job went missing.*":              "Active job went missing",
	"invalid timeZone.*":                     "Invalid timeZone",
	"unparseable schedule:.*":                "Unparseable schedule",
	"invalid schedule:.*":                    "Invalid schedule",
	"Missed scheduled time to start a job.*": "Missed scheduled time to start a job",
	"Get job.*":                              "Get job",
	"Error creating job:.*":                  "Error creating job",
}
