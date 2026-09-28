package aggregation

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Rules from the README example. The messages below are shortened copies of events rejected by
// VictoriaMetrics because the message label exceeded -maxLabelValueLen.
var exampleConfig = &Config{
	Rules: []Rule{
		// Spark Operator
		{Kind: "^SparkApplication$", Message: "^failed to submit SparkApplication ", Value: "Failed to submit SparkApplication"},
		{Kind: "^SparkApplication$", Message: `^SparkApplication \S+ failed: `, Value: "SparkApplication failed"},
		// NRM operator
		{Message: "^reconciliation error: ", Value: "Reconciliation error"},
		{Message: `^failed to execute reconciliation action \[(\w+)`, Value: "Failed to execute reconciliation action ${1}"},
		{Message: "^Failed to reconcile database users for services", Value: "Failed to reconcile database users for services"},
		{Message: "^Failed to reconcile database user ", Value: "Failed to reconcile database user"},
		{Message: "(?i)^reconciliation failed for (database|dataset) '", Value: "Reconciliation failed for ${1}"},
		{Message: "^Unable to execute GRANT queries for user ", Value: "Unable to execute GRANT queries for user"},
		{Message: "^Failed to grant privileges on schema ", Value: "Failed to grant privileges on schema"},
		{Message: `^Failed to get cassandra admin connection for database (\S+) `, Value: "Failed to get Cassandra admin connection for database ${1}"},
		{Message: "^it is not possible to modify existing PVC's '", Value: "It is not possible to modify existing PVC spec"},
		{Kind: "^Kafkatopic", Message: "^failed to connect to Kafka using url ", Value: "Failed to connect to Kafka"},
		{Kind: "^Kafkatopic", Message: `^Integration Kafka topic \S+ does not exist`, Value: "Integration Kafka topic does not exist"},
		// Core platform operators
		{Kind: "^Mesh$", Message: `^ErrCodeError \[([^\]]+)\]`, Value: "ErrCodeError ${1}"},
		{Kind: "^MaaS$", Message: "^failed to apply custom resource request: ([A-Za-z ]+)", Value: "Failed to apply custom resource request: ${1}"},
		{Kind: "^CDN$", Message: "^failed to create bucket ", Value: "Failed to create bucket"},
		{Kind: "^CDN$", Message: "^error during upload resource ", Value: "Failed to upload resource to bucket"},
		{Message: "^dbaas-aggregator rejected request: ", Value: "dbaas-aggregator rejected request"},
		// Argo CD
		{Kind: "^Application$", Message: "^Unable to delete application resources: ", Value: "Unable to delete application resources"},
		{Kind: "^Application$", Message: `^Sync operation to \S* ?failed`, Value: "Sync operation failed"},
		// OpenShift etcd operator
		{Reason: "^EtcdLeaderChangeMetrics$", Message: "^Detected leader change increase ", Value: "Detected leader change increase"},
	},
}

// initAggregations applies the config for one test and restores the defaults afterwards.
func initAggregations(t *testing.T, config *Config) {
	t.Helper()
	if err := InitAggregations(config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := InitAggregations(nil); err != nil {
			t.Errorf("failed to restore defaults: %v", err)
		}
	})
}

func TestGetCommonMessageWithCustomRules(t *testing.T) {
	initAggregations(t, exampleConfig)

	tests := []struct {
		kind, reason, message, expected string
	}{
		{
			"SparkApplication", "SparkApplicationSubmissionFailed",
			"failed to submit SparkApplication spark-app: failed to run spark-submit for SparkApplication spark-apps/spark-app: 26/09/24 12:03:58 WARN NativeCodeLoader: Unable to load native-hadoop library\n26/09/24 12:04:01 INFO SparkKubernetesClientFactory: Auto-configuring K8S client",
			"Failed to submit SparkApplication",
		},
		{
			"DatabaseUser-test-ns", "CustomUserCreation",
			`Unable to execute GRANT queries for user Ab1cD-test-user, database rb, version 1a2b3c4d-test: SQL query( GRANT USAGE ON SCHEMA "costed_events" TO "Ab1cD-test-user";`,
			"Unable to execute GRANT queries for user",
		},
		{
			"DatabaseUser-test-ns", "ExecuteReconciliationAction",
			`failed to execute reconciliation action [executeCustomGrantQueryAction: user: Xy2zW-test-rw-user, version: 1a2b3c4d-test, DB: rb]: SQL query( GRANT USAGE ON SCHEMA "geneva_data"`,
			"Failed to execute reconciliation action executeCustomGrantQueryAction",
		},
		{
			"ApplicationVersion-test-ns", "ReconcilingServices",
			"(combined from similar events): Failed to reconcile database users for services: following errors are occurred: [failed to reconcile rb database users for services: failed to connect to db rb",
			"Failed to reconcile database users for services",
		},
		{
			"ApplicationVersion-test-ns", "ReconcileVersion",
			"reconciliation error: %!w(*errors.errorString=&{following errors are occurred: [failed to prepare actions for Databases subreconciler",
			"Reconciliation error",
		},
		{
			"Mesh", "Failed to apply one or more configs",
			"ErrCodeError [CORE-MESH-CP-2002][2f009933-eee3-4ccf-a9d8-8677b0a87b3f] failed to register routes via v3 api: Executing actions in write transaction caused error",
			"ErrCodeError CORE-MESH-CP-2002",
		},
		{
			"Deployment", "EtcdLeaderChangeMetrics",
			`Detected leader change increase of 2.140548408502258 over 5 minutes on "OpenStack"; disk metrics are: etcd-master-2=0.115437`,
			"Detected leader change increase",
		},
		// The reason regexp does not match, so the message is only truncated.
		{
			"Deployment", "OtherReason",
			"Detected leader change increase of 2",
			"Detected leader change increase of 2",
		},
	}
	for _, tt := range tests {
		if got := GetCommonMessage(tt.kind, tt.reason, tt.message); got != tt.expected {
			t.Errorf("GetCommonMessage(%q, %q, ...) = %q, want %q", tt.kind, tt.reason, got, tt.expected)
		}
	}
}

func TestCustomRulesTakePrecedenceOverBuiltIn(t *testing.T) {
	initAggregations(t, &Config{Rules: []Rule{{Kind: "^Pod$", Message: "^Liveness probe failed", Value: "custom"}}})

	if got := GetCommonMessage("Pod", "Unhealthy", "Liveness probe failed: timeout"); got != "custom" {
		t.Fatalf("expected custom rule to win, got %q", got)
	}
}

func TestBuiltInPatternsHandleLongKubernetesMessages(t *testing.T) {
	initAggregations(t, nil)

	tests := []struct {
		kind, reason, message, expected string
	}{
		{
			"Service", "SyncLoadBalancerFailed",
			`(combined from similar events): Error syncing load balancer: failed to ensure load balancer: error creating loadbalancer {"name":"kube_service_kubernetes_ingress-nginx_ingress-nginx-controller"}`,
			"failed to ensure load balancer",
		},
		{
			"Pod", "ProbeError",
			"Readiness probe error: HTTP probe failed with statuscode: 500\nbody: [+]ping ok\n[+]log ok\n[-]etcd failed: reason withheld",
			"Readiness probe error",
		},
		{
			"Pod", "Unhealthy",
			"Readiness probe errored: rpc error: code = Unknown desc = command timed out",
			"Readiness probe failed",
		},
		{
			"Pod", "FailedCreatePodSandBox",
			`Failed to create pod sandbox: rpc error: code = Unknown desc = failed to setup network for sandbox "35b9ce37c2086add88c3d4c28db0da9f": plugin type="calico" failed (add)`,
			"Failed to create pod sandbox",
		},
		{
			"Pod", "FailedMount",
			"Unable to attach or mount volumes: unmounted volumes=[nifi-pvol-data nifi-prov-data], unattached volumes=[nifi-prov-data nifi-pvol-data], failed to process volumes=[]: timed out waiting for the condition",
			"Unable to attach or mount volumes",
		},
		{
			"Pod", "FailedScheduling",
			"0/27 nodes are available: 17 node(s) didn't match pod topology spread constraints (missing required label), 3 node(s) had untolerated taint(s), 7 Too many pods.",
			"Nodes are not available for scheduling",
		},
		// A specific pattern wins over the generic one that matches at the same position.
		{
			"Pod", "FailedScheduling",
			"0/3 nodes are available: 3 node(s) had untolerated taint {node-role.kubernetes.io/master: }.",
			"Nodes had untolerated taint",
		},
		{
			"Pod", "Evicted",
			"The node was low on resource: ephemeral-storage. Threshold quantity: 14484819724, available: 13373152Ki.",
			"The node was low on resource: ephemeral-storage",
		},
		{
			"HorizontalPodAutoscaler", "FailedGetResourceMetric",
			`failed to get cpu utilization: unable to get metrics for resource cpu: unable to fetch metrics from resource metrics API: pods.metrics.k8s.io is forbidden`,
			"Failed to get cpu utilization",
		},
		{
			"HorizontalPodAutoscaler", "FailedGetPodsMetric",
			`unable to get metric tm_busyness: unable to fetch metrics from custom metrics API: pods.custom.metrics.k8s.io "*" is forbidden`,
			"Unable to get metric tm_busyness",
		},
		{
			"HorizontalPodAutoscaler", "FailedGetExternalMetric",
			"unable to get external metric test-ns/s0-kafka-incoming-messages/&LabelSelector{MatchLabels:map[string]string{scaledobject.keda.sh/name: processor,},}: unable to fetch metrics",
			"Unable to get external metric s0-kafka-incoming-messages",
		},
		{
			"PodDisruptionBudget", "UnmanagedPods",
			"Pods selected by this PodDisruptionBudget (selector: &LabelSelector{MatchLabels:map[string]string{app: consul,},}) were found to be unmanaged or managed by an unsupported controller",
			"Found unmanaged pods associated with this PDB",
		},
		{
			"StatefulSet", "FailedCreate",
			`Create Pod web-ui-0 in StatefulSet web-ui failed error: Pod "web-ui-0" is invalid: metadata.labels: Invalid value`,
			"Pod configuration is invalid",
		},
		{
			"Node", "FreeDiskSpaceFailed",
			"Insufficient free disk space on the node's image filesystem (85% of 89.9 GiB used). Failed to free sufficient space by deleting unused images (freed 36018163 bytes).",
			"Insufficient free disk space on the node's image filesystem",
		},
		{
			"Node", "ImageGCFailed",
			"(combined from similar events): wanted to free 1258866278 bytes, but freed 0 bytes space with errors in image deletion: rpc error: code = Unknown",
			"Failed to free disk space because of errors in image deletion",
		},
		// Kubernetes API errors are recognized for any kind.
		{
			"Report", "ExecuteReconciliationAction",
			`failed to execute reconciliation action [UpdateResourceAction: ConfigMap: cache]: Operation cannot be fulfilled on configmaps "cache": the object has been modified; please apply your changes to the latest version and try again`,
			"Operation cannot be fulfilled: the object has been modified",
		},
		{
			"Ingress", "FailedIngressToRouteConversion",
			`Error in converting Ingress to Route: Operation cannot be fulfilled on ingresses.networking.k8s.io "ea-api-ea": StorageError: invalid object, Code: 4`,
			"Operation cannot be fulfilled: invalid object",
		},
		{
			"VMAuth", "ReconciliationError",
			`cannot create or update vmauth deploy: deployments.apps is forbidden: User "system:serviceaccount:monitoring:vm-operator" cannot create resource "deployments" in API group "apps"`,
			"Forbidden: User cannot create resource deployments",
		},
		{
			"SparkApplication", "SparkApplicationFailed",
			`SparkApplication app-k7bnrckfmvw08kpu failed: failed to sync PodGroup with error: admission webhook "validatepodgroup.volcano.sh" denied the request: unable to find queue`,
			"Admission webhook validatepodgroup.volcano.sh denied the request",
		},
		{
			"ClusterPolicy", "PolicyError",
			`policy generate-limitrange/generate-limitrange error: Internal error occurred: failed calling webhook "validate.kyverno.svc-fail": failed to call webhook`,
			"Failed calling webhook validate.kyverno.svc-fail",
		},
		{
			"Pod", "Unhealthy",
			"Liveness probe failed: [2026-09-26 10:22:32,226] [INFO] [settings.py] [thread=MainThread] Configured default timezone UTC\nUsage: python -m celery [OPTIONS] COMMAND [ARGS]...\n\nError: Invalid value for '-A' / '--app': \nUnable to load celery application.",
			"Liveness probe failed",
		},
	}
	for _, tt := range tests {
		if got := GetCommonMessage(tt.kind, tt.reason, tt.message); got != tt.expected {
			t.Errorf("GetCommonMessage(%q, %q, ...) = %q, want %q", tt.kind, tt.reason, got, tt.expected)
		}
	}
}

func TestUnmatchedMessageIsTruncated(t *testing.T) {
	initAggregations(t, &Config{MaxMessageLength: 20})

	if got := GetCommonMessage("UnknownKind", "Failed", "first line\nsecond line"); got != "first line" {
		t.Errorf("expected only the first line, got %q", got)
	}
	if got := GetCommonMessage("UnknownKind", "Failed", "(combined from similar events): short"); got != "short" {
		t.Errorf("expected the correlator prefix to be removed, got %q", got)
	}
	got := GetCommonMessage("UnknownKind", "Failed", strings.Repeat("ж", 50))
	if utf8.RuneCountInString(got) != 20 || !strings.HasSuffix(got, truncationSuffix) || !utf8.ValidString(got) {
		t.Errorf("expected a valid 20-character value ending with %q, got %q", truncationSuffix, got)
	}
}

func TestDefaultMaxMessageLength(t *testing.T) {
	initAggregations(t, nil)

	got := GetCommonMessage("UnknownKind", "Failed", strings.Repeat("a", 5000))
	if len(got) != DefaultMaxMessageLength {
		t.Errorf("expected %d characters, got %d", DefaultMaxMessageLength, len(got))
	}
}

func TestApplyConfigRejectsInvalidRules(t *testing.T) {
	initAggregations(t, nil)

	tests := []*Config{
		{MaxMessageLength: -1},
		{Rules: []Rule{{Message: "", Value: "v"}}},
		{Rules: []Rule{{Message: "m", Value: ""}}},
		{Rules: []Rule{{Message: "(", Value: "v"}}},
		{Rules: []Rule{{Kind: "(", Message: "m", Value: "v"}}},
		{Rules: []Rule{{Reason: "(", Message: "m", Value: "v"}}},
	}
	for _, config := range tests {
		if err := InitAggregations(config); err == nil {
			t.Errorf("expected an error for %+v", config)
		}
	}
	if len(customRules) != 0 {
		t.Errorf("expected no custom rules after a failed config, got %d", len(customRules))
	}
}
