package aggregation

import "regexp"

var nodeAggregationRegexps = map[int]*regexp.Regexp{}
var nodeAggregationLabelValues = map[int]string{}
var nodeAggregations = map[string]string{
	".*The node was low on resource.*":                                                         "The node was low on resource",
	"Failed to update Node Allocatable Limits.*":                                               "Failed to update Node Allocatable Limits",
	"Failed to enforce System Reserved Cgroup Limits on.*":                                     "Failed to enforce System Reserved Cgroup Limits",
	"Failed to enforce Kube Reserved Cgroup Limits on .*":                                      "Failed to enforce Kube Reserved Cgroup Limits",
	"Resolv.conf file .* contains search line consisting of more than \\d+ domains!":           "Resolv.conf file contains search line consisting of more than domain count limit!",
	"Resolv.conf file .* contains a search path which length is more than allowed \\d+ chars!": "Resolv.conf file contains a search path which length is more than allowed subdomain length!",
	"Resolv.conf file .* contains search line which length is more than allowed \\d+ chars!":   "Resolv.conf file contains search line which length is more than max number of characters in the search path",
}
