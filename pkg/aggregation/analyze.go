package aggregation

import (
	"regexp"
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

func GetCommonMessage(kind string, reason string, message string) string {
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

func InitAggregations() {
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
}

func initAggregationsForKind(aggregations map[string]string, regexps map[int]*regexp.Regexp, labelValues map[int]string) {
	it := 0
	for expression, value := range aggregations {
		regexps[it] = regexp.MustCompile(expression)
		labelValues[it] = value
		it++
	}
	clear(aggregations)
}

func getCommonMessageForEvent(reason string, message string, regexps map[int]*regexp.Regexp, labelValues map[int]string) string {
	for index, expression := range regexps {
		if expression.MatchString(message) {
			return labelValues[index]
		}
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
