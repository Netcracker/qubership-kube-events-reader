package aggregation

import "regexp"

var serviceAggregationRegexps = map[int]*regexp.Regexp{}
var serviceAggregationLabelValues = map[int]string{}

var serviceAggregations = map[string]string{
	"Error listing Pods for Service.*":                         "Error listing Pods for Service",
	"Error listing Endpoint Slices for Service.*":              "Error listing Endpoint Slices for Service",
	"Error updating Endpoint Slices for Service.*":             "Error updating Endpoint Slices for Service",
	"failed to check if load balancer exists before cleanup.*": "failed to check if load balancer exists before cleanup",
	"failed to delete load balancer.*":                         "failed to delete load balancer",
	"failed to remove load balancer cleanup finalizer.*":       "failed to remove load balancer cleanup finalizer",
	"failed to add load balancer cleanup finalizer.*":          "failed to add load balancer cleanup finalizer",
	"failed to ensure load balancer.*":                         "failed to ensure load balancer",
	"failed to update load balancer status.*":                  "failed to update load balancer status",
	"Error updating load balancer with new hosts.*":            "Error updating load balancer with new hosts",
	"Error deleting load balancer.*":                           "Error deleting load balancer",
}

var endpointsAggregationRegexps = map[int]*regexp.Regexp{}
var endpointsAggregationLabelValues = map[int]string{}

var endpointsAggregations = map[string]string{
	"Failed to create endpoint for service.*":                            "Failed to create endpoint for service",
	"Failed to update endpoint.*":                                        "Failed to update endpoint",
	"Skipped \\d+ invalid IP addresses when mirroring to EndpointSlices": "Skipped invalid IP addresses when mirroring to EndpointSlices",
	"A max of \\d+ addresses can be mirrored to EndpointSlices per Endpoints subset. \\d+ addresses were skipped": "Addresses in Endpoints were skipped due to exceeding MaxEndpointsPerSubset",
}
