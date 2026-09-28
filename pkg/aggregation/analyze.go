package aggregation

import (
	"cmp"
	"maps"
	"regexp"
	"slices"
	"strings"
)

const (
	kindPod                   = "pod"
	kindPodDisruptionBudget   = "poddisruptionbudget"
	kindDaemonSet             = "daemonset"
	kindReplicaSet            = "replicaset"
	kindReplicationController = "replicationcontroller"
	kindDeployment            = "deployment"
	kindDeploymentConfig      = "deploymentconfig"
	kindGrafanaDashboard      = "grafanadashboard"
	kindPVC                   = "persistentvolumeclaim"
	kindPV                    = "persistentvolume"
	kindHPA                   = "horizontalpodautoscaler"
	kindNode                  = "node"
	kindStatefulSet           = "statefulset"
	kindClusterIssuer         = "clusterissuer"
	kindIssuer                = "issuer"
	kindChallenge             = "challenge"
	kindCSR                   = "certificatesigningrequest"
	kindCertificate           = "certificate"
	kindOrder                 = "order"
	kindService               = "service"
	kindEndpoints             = "endpoints"
	kindJob                   = "job"
	kindCronJob               = "cronjob"
)

var (
	forbiddenMessageRegexp     = regexp.MustCompile(".*is forbidden: User .* cannot update resource .* in API group")
	ownerRefDoesNotExistRegexp = regexp.MustCompile("ownerRef .* does not exist in namespace.*")
)

// apiErrorAggregations match errors of the Kubernetes API server that any controller can put into an event message.
// They are checked in order for all kinds when neither custom rules nor kind patterns match.
var apiErrorAggregations = []struct {
	expression *regexp.Regexp
	value      string
}{
	{regexp.MustCompile("Operation cannot be fulfilled on .*the object has been modified"), "Operation cannot be fulfilled: the object has been modified"},
	{regexp.MustCompile("Operation cannot be fulfilled on .*StorageError: invalid object"), "Operation cannot be fulfilled: invalid object"},
	{regexp.MustCompile("Operation cannot be fulfilled on "), "Operation cannot be fulfilled"},
	{regexp.MustCompile(`is forbidden: [Uu]ser "[^"]*"(?: \(groups=[^)]*\))? cannot ([\w-]+) resource "([^"]+)"`), "Forbidden: User cannot ${1} resource ${2}"},
	{regexp.MustCompile(`admission webhook "([^"]+)" denied the request`), "Admission webhook ${1} denied the request"},
	{regexp.MustCompile(`failed calling webhook "([^"]+)"`), "Failed calling webhook ${1}"},
}

// GetCommonMessage returns a low-cardinality value for the message label.
// Custom rules are checked first, then the built-in patterns for the kind.
// A message that matches nothing is cut to its first line and to the maximum message length.
func GetCommonMessage(kind string, reason string, message string) string {
	message = strings.TrimPrefix(message, combinedEventsPrefix)
	if value, ok := getMessageByCustomRules(kind, reason, message); ok {
		return truncateMessage(value)
	}
	value := getBuiltInMessage(kind, reason, message)
	if value == message {
		value = getMessageByAPIError(message)
	}
	return truncateMessage(value)
}

func getMessageByAPIError(message string) string {
	for _, aggregation := range apiErrorAggregations {
		if match := aggregation.expression.FindStringSubmatchIndex(message); match != nil {
			return string(aggregation.expression.ExpandString(nil, aggregation.value, message, match))
		}
	}
	return message
}

func getBuiltInMessage(kind string, reason string, message string) string {
	switch strings.ToLower(kind) {
	case kindPod:
		return getCommonMessageForEvent(reason, message, podAggregationRegexps, podAggregationLabelValues)
	case kindPodDisruptionBudget:
		return getCommonMessageForEvent(reason, message, podDisruptionBudgetAggregationRegexps, podDisruptionBudgetAggregationLabelValues)
	case kindDaemonSet:
		return getCommonMessageForEvent(reason, message, dsAggregationRegexps, dsAggregationLabelValues)
	case kindReplicaSet, kindReplicationController:
		return getCommonMessageForEvent(reason, message, rsAggregationRegexps, rsAggregationLabelValues)
	case kindDeployment:
		return getCommonMessageForEvent(reason, message, depAggregationRegexps, depAggregationLabelValues)
	case kindGrafanaDashboard:
		return getCommonMessageForEvent(reason, message, grafanaAggregationRegexps, grafanaAggregationLabelValues)
	case kindPVC:
		return getCommonMessageForEvent(reason, message, pvcAggregationRegexps, pvcAggregationLabelValues)
	case kindPV:
		return getCommonMessageForEvent(reason, message, pvAggregationRegexps, pvAggregationLabelValues)
	case kindHPA:
		return getCommonMessageForHPA(reason, message, hpaAggregationRegexps, hpaAggregationLabelValues)
	case kindNode:
		return getCommonMessageForEvent(reason, message, nodeAggregationRegexps, nodeAggregationLabelValues)
	case kindStatefulSet:
		return getCommonMessageForEvent(reason, message, ssAggregationRegexps, ssAggregationLabelValues)
	case kindClusterIssuer, kindIssuer:
		return getCommonMessageForCertManager(reason, message)
	case kindChallenge:
		return getCommonMessageForEvent(reason, message, challengeAggregationRegexps, challengeAggregationLabelValues)
	case kindCSR:
		return getCommonMessageForEvent(reason, message, csrAggregationRegexps, csrAggregationLabelValues)
	case kindCertificate:
		return getCommonMessageForEvent(reason, message, certificateAggregationRegexps, certificateAggregationLabelValues)
	case kindOrder:
		return getCommonMessageForEvent(reason, message, orderAggregationRegexps, orderAggregationLabelValues)
	case kindDeploymentConfig:
		return getCommonMessageForEvent(reason, message, depConfigAggregationRegexps, depConfigAggregationLabelValues)
	case kindService:
		return getCommonMessageForEvent(reason, message, serviceAggregationRegexps, serviceAggregationLabelValues)
	case kindEndpoints:
		return getCommonMessageForEvent(reason, message, endpointsAggregationRegexps, endpointsAggregationLabelValues)
	case kindJob:
		return getCommonMessageForEvent(reason, message, jobAggregationRegexps, jobAggregationLabelValues)
	case kindCronJob:
		return getCommonMessageForEvent(reason, message, cronjobAggregationRegexps, cronjobAggregationLabelValues)
	default:
		return getAllKindsMessageByReason(reason, message)
	}
}

// InitAggregations compiles the built-in patterns and applies the user-defined configuration.
// The configuration can be nil.
func InitAggregations(config *Config) error {
	initAggregationsForKind(podAggregations, podAggregationRegexps, podAggregationLabelValues)
	initAggregationsForKind(podDisruptionBudgetAggregations, podDisruptionBudgetAggregationRegexps, podDisruptionBudgetAggregationLabelValues)
	initAggregationsForKind(dsAggregations, dsAggregationRegexps, dsAggregationLabelValues)
	initAggregationsForKind(depAggregations, depAggregationRegexps, depAggregationLabelValues)
	initAggregationsForKind(depConfigAggregations, depConfigAggregationRegexps, depConfigAggregationLabelValues)
	initAggregationsForKind(rsAggregations, rsAggregationRegexps, rsAggregationLabelValues)
	initAggregationsForKind(grafanaAggregations, grafanaAggregationRegexps, grafanaAggregationLabelValues)
	initAggregationsForKind(pvcAggregations, pvcAggregationRegexps, pvcAggregationLabelValues)
	initAggregationsForKind(pvAggregations, pvAggregationRegexps, pvAggregationLabelValues)
	initAggregationsForKind(ssAggregations, ssAggregationRegexps, ssAggregationLabelValues)
	initAggregationsForKind(nodeAggregations, nodeAggregationRegexps, nodeAggregationLabelValues)
	initAggregationsForKind(hpaAggregations, hpaAggregationRegexps, hpaAggregationLabelValues)
	initAggregationsForKind(serviceAggregations, serviceAggregationRegexps, serviceAggregationLabelValues)
	initAggregationsForKind(endpointsAggregations, endpointsAggregationRegexps, endpointsAggregationLabelValues)
	initAggregationsForKind(jobAggregations, jobAggregationRegexps, jobAggregationLabelValues)
	initAggregationsForKind(cronjobAggregations, cronjobAggregationRegexps, cronjobAggregationLabelValues)
	initAggregationsForKind(issuerAggregations, issuerAggregationRegexps, issuerAggregationLabelValues)
	initAggregationsForKind(orderAggregations, orderAggregationRegexps, orderAggregationLabelValues)
	initAggregationsForKind(csrAggregations, csrAggregationRegexps, csrAggregationLabelValues)
	initAggregationsForKind(certificateAggregations, certificateAggregationRegexps, certificateAggregationLabelValues)
	initAggregationsForKind(challengeAggregations, challengeAggregationRegexps, challengeAggregationLabelValues)
	return applyConfig(config)
}

func initAggregationsForKind(aggregations map[string]string, regexps map[int]*regexp.Regexp, labelValues map[int]string) {
	// Longer patterns go first, so getCommonMessageForEvent prefers a specific pattern over a generic one
	// that matches at the same position. The order is stable between runs.
	expressions := slices.SortedFunc(maps.Keys(aggregations), func(a, b string) int {
		return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b))
	})
	for it, expression := range expressions {
		regexps[it] = regexp.MustCompile(expression)
		labelValues[it] = aggregations[expression]
	}
	clear(aggregations)
}

func getCommonMessageForEvent(reason string, message string, regexps map[int]*regexp.Regexp, labelValues map[int]string) string {
	// Several patterns can match one message, for example "Liveness probe failed" and "Error: .*" in the probe output.
	// The match that starts first wins. On a tie the longer pattern wins, because patterns are sorted by length.
	best, bestMatch := -1, []int(nil)
	for index := range len(regexps) {
		match := regexps[index].FindStringSubmatchIndex(message)
		if match != nil && (bestMatch == nil || match[0] < bestMatch[0]) {
			best, bestMatch = index, match
			if match[0] == 0 {
				break
			}
		}
	}
	if best >= 0 {
		return string(regexps[best].ExpandString(nil, labelValues[best], message, bestMatch))
	}
	if strings.EqualFold(reason, "OwnerRefInvalidNamespace") {
		return "ownerRef does not exist in namespace"
	}
	return message
}

func getCommonMessageForHPA(reason string, message string, regexps map[int]*regexp.Regexp, labelValues map[int]string) string {
	if strings.EqualFold(reason, "FailedGetScale") || strings.EqualFold(reason, "FailedComputeMetricsReplicas") {
		return reason
	}
	return getCommonMessageForEvent(reason, message, regexps, labelValues)
}

func getAllKindsMessageByReason(reason string, message string) string {
	if strings.EqualFold(reason, "OwnerRefInvalidNamespace") {
		if ownerRefDoesNotExistRegexp.MatchString(message) {
			return "ownerRef does not exist in namespace"
		}
	}
	if strings.EqualFold(reason, "UpdateError") {
		if forbiddenMessageRegexp.MatchString(message) {
			return "Forbidden: User cannot update resource"
		}
	}
	return message
}
