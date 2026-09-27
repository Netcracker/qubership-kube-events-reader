package aggregation

import "regexp"

var ssAggregationRegexps = map[int]*regexp.Regexp{}
var ssAggregationLabelValues = map[int]string{}
var ssAggregations = map[string]string{
	"[Dd]elete Pod .* failed error.*":                                      "Delete Pod error",
	"[Dd]elete Pod .* successful":                                          "Delete Pod successful",
	"[Cc]reate Pod .* successful":                                          "Create Pod successful",
	"StatefulSet .* is recreating failed Pod .*":                           "StatefulSet is recreating failed Pod",
	"StatefulSet .* is recreating terminated Pod .*":                       "StatefulSet is recreating terminated Pod",
	"[Cc]reate Pod .* in StatefulSet .* failed error: Pod .* is invalid.*": "Pod configuration is invalid",
	"[Cc]reate Pod .* in StatefulSet .* failed error: pods .* forbidden.*": "Pod is forbidden",
	"PersistentVolumeClaim .* has a conflicting OwnerReference that acts as a manging controller, the retention policy is ignored for this claim.*": "PersistentVolumeClaim has a conflicting OwnerReference",
	".*create Claim .* Pod .* in StatefulSet .* success.*":          "Create Claim for Pod in StatefulSet success",
	".*create Claim .* for Pod .* in StatefulSet .* failed error.*": "Create Claim for Pod in StatefulSet failed error",
}
