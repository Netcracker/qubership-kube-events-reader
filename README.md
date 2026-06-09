# Qubership kube events reader

This component is used for collecting Kubernetes event logs from Cloud.

## Table of Content

* [Qubership kube events reader](#qubership-kube-events-reader)
  * [Table of Content](#table-of-content)
  * [Overview](#overview)
    * [Command line arguments](#command-line-arguments)
    * [Events metrics](#events-metrics)
    * [Event log example](#event-log-example)
  * [Repository structure](#repository-structure)
  * [How to start](#how-to-start)
    * [Build](#build)
    * [Definition of done](#definition-of-done)
    * [Deploy](#deploy)
      * [Prerequisites](#prerequisites)
      * [HWE and Limits](#hwe-and-limits)
      * [Deploy with helm](#deploy-with-helm)
    * [How to debug](#how-to-debug)
    * [How to troubleshoot](#how-to-troubleshoot)

## Overview

K8s events Reader is a deployment that observes for Kubernetes events and send its to configured output. Now it
supports two types of output: print events to logs in predefined format (to
be processed by Fluentd/FluentBit) or/and collect events as metrics (and provide endpoint to scrape metrics). It is
deployed as a part of cloud Logging and Monitoring stacks.

It implements Kubernetes controller that watches for kind Event with API version events.k8s.io/v1 adding and modifying
and collect or send data of event.

### Command line arguments

Entrypoint of K8s events Reader is `/events-reader/eventsreader`.

<!-- markdownlint-disable line-length -->

| Argument      | Default value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Description                                                                                                                                  |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `namespace`   | `-`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Namespace to watch for events. The parameter can be used multiple times.<br>If parameter is not set events of all namespaces will be watched |
| `output`      | `logs`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Outputs for events. The parameter can be used multiple times. The parameter has two available values: metrics or/and logs                    |
| `metricsPort` | `9999`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Port to expose Prometheus metrics on                                                                                                         |
| `metricsPath` | `/metrics`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | HTTP path to scrape for Prometheus metrics                                                                                                   |
| `filtersPath` | `-`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Absolute path to file with filter events configuration                                                                                       |
| `format`      | <details><summary>value</summary>{\"time\":\"{{.LastTimestamp.Format \"2006-01-02T15:04:05Z\"}}\",\"involvedObjectKind\":\"{{.InvolvedObject.Kind}}\",\"involvedObjectNamespace\":\"{{.InvolvedObject.Namespace}}\",\"involvedObjectName\":\"{{.InvolvedObject.Name}}\",\"involvedObjectUid\":\"{{.InvolvedObject.UID}}\",\"involvedObjectApiVersion\":\"{{.InvolvedObject.APIVersion}}\",\"involvedObjectResourceVersion\":\"{{.InvolvedObject.ResourceVersion}}\",\"reason\":\"{{.Reason}}\",\"type\":\"{{.Type}}\",\"message\":\"{{js .Message}}\",\"kind\":\"KubernetesEvent\"}</details> | Format to print Event. It should be valid Golang template of `text/template` package                                                         |
| `workers`     | `2`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Workers number for controller                                                                                                                |
| `pprofEnable` | `true`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Enable pprof                                                                                                                                 |
| `pprofAddr`   | `8080`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Port to health and pprof endpoint                                                                                                            |

<!-- markdownlint-enable line-length -->

Example:

<!-- markdownlint-disable line-length -->

```shell
/events-reader/eventsreader -workers=2 -output=logs -namespace=logging -namespace=monitoring -format={"time":"{{.LastTimestamp.Format "2006-01-02T15:04:05Z"},"name":"{{.InvolvedObject.Name}}"}"
```

<!-- markdownlint-enable line-length -->

**Note**: You need to escape value of `format` parameter when you set value in `cloudEventsReader.args`. In pod `args`
it should be set as it is, because K8s events Reader read it as simple string (not as json), as you can set format
not only to be printed in json format.

### Events metrics

When you run qubership-kube-events-reader with `-output=metrics` the application will collect the next list of metrics:

<!-- markdownlint-disable line-length -->

| Metric                                            | Type    | Labels                                                                                | Description                                                        |
|---------------------------------------------------|---------|---------------------------------------------------------------------------------------|--------------------------------------------------------------------|
| `kube_events_total`                              | counter | kind, event_namespace, type                                                           | Count of kubernetes events                                         |
| `kube_events_normal_total`                       | counter | kind, event_object, event_namespace, reason, controller, controller_instance, message | Count of kubernetes events with type normal aggregated by message  |
| `kube_events_warning_total`                      | counter | kind, event_object, event_namespace, reason, controller, controller_instance, message | Count of kubernetes events with type warning aggregated by message |
| `kube_events_reporting_controller_normal_total`  | counter | controller, controller_instance, kind, event_namespace                                | Count of kubernetes events with type normal                        |
| `kube_events_reporting_controller_warning_total` | counter | controller, controller_instance, kind, event_namespace                                | Count of kubernetes events with type warning                       |

The example of events metrics:

```text
# HELP kube_events_total Count of kubernetes events
# TYPE kube_events_total counter
kube_events_total{kind="Certificate",event_namespace="test-ns1",type="Normal"} 4
kube_events_total{kind="Certificate",event_namespace="test-ns3",type="Normal"} 8
kube_events_total{kind="CertificateRequest",event_namespace="test-ns1",type="Normal"} 7
kube_events_total{kind="CertificateRequest",event_namespace="test-ns3",type="Normal"} 14
kube_events_total{kind="ClusterIssuer",event_namespace="",type="Warning"} 500
kube_events_total{kind="ConfigMap",event_namespace="ingress-nginx",type="Normal"} 2
kube_events_total{kind="CronJob",event_namespace="default",type="Normal"} 8
kube_events_total{kind="DaemonSet",event_namespace="test-ns2",type="Warning"} 156
kube_events_total{kind="Deployment",event_namespace="kafka-service",type="Normal"} 120
kube_events_total{kind="Deployment",event_namespace="monitoring",type="Normal"} 2
kube_events_total{kind="Deployment",event_namespace="test-ns1",type="Normal"} 2
kube_events_total{kind="Deployment",event_namespace="test-ns3",type="Normal"} 4
kube_events_total{kind="Endpoints",event_namespace="kafka-service",type="Warning"} 2
kube_events_total{kind="Endpoints",event_namespace="test-ns1",type="Warning"} 4
kube_events_total{kind="Endpoints",event_namespace="test-ns3",type="Warning"} 13
kube_events_total{kind="GrafanaDashboard",event_namespace="mistral-vaze",type="Normal"} 2
kube_events_total{kind="GrafanaDashboard",event_namespace="monitoring",type="Normal"} 1
kube_events_total{kind="GrafanaDashboard",event_namespace="test-ns3",type="Normal"} 5
kube_events_total{kind="GrafanaDashboard",event_namespace="test-ns3",type="Warning"} 1
kube_events_total{kind="Ingress",event_namespace="arango-ek",type="Normal"} 14
kube_events_total{kind="Ingress",event_namespace="cassandra",type="Normal"} 14
kube_events_total{kind="Ingress",event_namespace="mongodb",type="Normal"} 14
kube_events_total{kind="Ingress",event_namespace="test-ns3",type="Normal"} 19
kube_events_total{kind="Job",event_namespace="default",type="Normal"} 4
kube_events_total{kind="Job",event_namespace="mistral-vaze",type="Normal"} 2
kube_events_total{kind="Job",event_namespace="test-ns1",type="Normal"} 8
kube_events_total{kind="Job",event_namespace="test-ns3",type="Normal"} 8
kube_events_total{kind="MistralService",event_namespace="mistral-vaze",type="Error"} 1
kube_events_total{kind="MistralService",event_namespace="mistral-vaze",type="Normal"} 1
kube_events_total{kind="PersistentVolumeClaim",event_namespace="test-ns4",type="Warning"} 1500
kube_events_total{kind="PersistentVolumeClaim",event_namespace="clickhouse-test",type="Normal"} 500
kube_events_total{kind="PersistentVolumeClaim",event_namespace="kafka-service",type="Normal"} 19
kube_events_total{kind="Pod",event_namespace="default",type="Normal"} 8
kube_events_total{kind="Pod",event_namespace="ingress-nginx",type="Normal"} 40
kube_events_total{kind="Pod",event_namespace="ingress-nginx",type="Warning"} 2
kube_events_total{kind="Pod",event_namespace="postgres",type="Normal"} 113
kube_events_total{kind="Pod",event_namespace="postgres",type="Warning"} 1
kube_events_total{kind="PodDisruptionBudget",event_namespace="test-ns1",type="Normal"} 8
kube_events_total{kind="PodDisruptionBudget",event_namespace="test-ns1",type="Warning"} 5
kube_events_total{kind="ReplicaSet",event_namespace="test-ns3",type="Normal"} 4
kube_events_total{kind="StatefulSet",event_namespace="postgres",type="Normal"} 55
kube_events_total{kind="StatefulSet",event_namespace="postgres",type="Warning"} 57
# HELP kube_events_normal_total Count of kubernetes events with type normal aggregated by message
# TYPE kube_events_normal_total counter
kube_events_normal_total{controller="",controller_instance="",kind="CertificateRequest",message="Certificate request has been approved by cert-manager.io",event_namespace="test-ns1",event_object="vault-service-server-tls-certificate-cf2ln",reason="cert-manager.io"} 1
kube_events_normal_total{controller="",controller_instance="",kind="GrafanaDashboard",message="dashboard successfully submitted",event_namespace="test-ns3",event_object="vault-service-grafana-dashboard",reason="Success"} 5
kube_events_normal_total{controller="controllermanager",controller_instance="",kind="PodDisruptionBudget",message="No matching pods found",event_namespace="ivsh-click",event_object="cluster-replicated",reason="NoPods"} 2
kube_events_normal_total{controller="controllermanager",controller_instance="",kind="PodDisruptionBudget",message="No matching pods found",event_namespace="test-ns1",event_object="vault-service",reason="NoPods"} 8
kube_events_normal_total{controller="cronjob-controller",controller_instance="",kind="CronJob",message="Created job",event_namespace="default",event_object="kube-cleanup-cronjob",reason="SuccessfulCreate"} 1
kube_events_normal_total{controller="default-scheduler",controller_instance="",kind="Pod",message="Successfully assigned",event_namespace="kafka-service",event_object="kafka-3-6f8c469785-78nsl",reason="Scheduled"} 1
kube_events_normal_total{controller="default-scheduler",controller_instance="",kind="Pod",message="Successfully assigned",event_namespace="kafka-service",event_object="kafka-3-6f8c469785-ccq64",reason="Scheduled"} 1
kube_events_normal_total{controller="kubelet",kind="Pod",controller_instance="10.0.2.15",message="Stopping container",event_namespace="dmsh-pg-test",event_object="patroni-core-operator-85689ccbd8-4kt7z",reason="Killing"} 1
kube_events_normal_total{controller="nginx-ingress-controller",controller_instance="",kind="Ingress",message="Scheduled for sync",event_namespace="test-ns3",event_object="vault-service",reason="Sync"} 19
kube_events_normal_total{controller="nginx-ingress-controller",controller_instance="",kind="Pod",message="NGINX reload triggered due to a change in configuration",event_namespace="ingress-nginx",event_object="ingress-nginx-controller-8tn9z",reason="RELOAD"} 10
kube_events_normal_total{controller="replicaset-controller",controller_instance="",kind="ReplicaSet",message="Created pod",event_namespace="dmsh-pg-test",event_object="patroni-core-operator-5597dd77c5",reason="SuccessfulCreate"} 1
kube_events_normal_total{controller="replicaset-controller",controller_instance="",kind="ReplicaSet",message="Created pod",event_namespace="dmsh-pg-test",event_object="patroni-core-operator-68c7dbd696",reason="SuccessfulCreate"} 1
kube_events_normal_total{controller="statefulset-controller",controller_instance="",kind="StatefulSet",message="Create Pod successful",event_namespace="test-ns3",event_object="vault-service",reason="SuccessfulCreate"} 6
kube_events_normal_total{controller="statefulset-controller",controller_instance="",kind="StatefulSet",message="StatefulSet is recreating terminated Pod",event_namespace="test-ns1",event_object="vault-service",reason="RecreatingTerminatedPod"} 30
# HELP kube_events_warning_total Count of kubernetes events with type warning aggregated by message
# TYPE kube_events_warning_total counter
kube_events_warning_total{controller="",controller_instance="",kind="ClusterIssuer",message="Error getting keypair for CA issuer",event_namespace="",event_object="dev-cluster-issuer",reason="ErrGetKeyPair"} 500
kube_events_warning_total{controller="",controller_instance="",kind="GrafanaDashboard",message="error creating folder, expected status 200 but got 409",event_namespace="test-ns1",event_object="vault-service-grafana-dashboard",reason="ProcessingError"} 1
kube_events_warning_total{controller="",controller_instance="",kind="Pod",message="Liveness probe failed",event_namespace="clickhouse-jyya",event_object="chi-cluster-replicated-0-0-0",reason="Unhealthy"} 4
kube_events_warning_total{controller="controllermanager",controller_instance="",kind="PodDisruptionBudget",message="Failed to calculate the number of expected pods",event_namespace="test-ns1",event_object="vault-service",reason="CalculateExpectedPodCountFailed"} 1
kube_events_warning_total{controller="endpoint-controller",controller_instance="",kind="Endpoints",message="Failed to update endpoint",event_namespace="kafka-service",event_object="kafka",reason="FailedToUpdateEndpoint"} 1
kube_events_warning_total{controller="kubelet",controller_instance="10.0.2.15",kind="Pod",message="Back-off restarting failed container in pod",event_namespace="kafka-service",event_object="kafka-2-8655b9f55c-lsc9z",reason="BackOff"} 9
kube_events_warning_total{controller="kubelet",controller_instance="10.0.2.15",kind="Pod",message="Nameserver limits were exceeded",event_namespace="kube-system",event_object="kube-proxy-m8ntz",reason="DNSConfigForming"} 500
kube_events_warning_total{controller="kubelet",controller_instance="10.0.2.15",kind="Pod",message="Volume is not found",event_namespace="test-ns1",event_object="vault-service-1",reason="FailedMount"} 4
kube_events_warning_total{controller="persistentvolume-controller",controller_instance="",kind="PersistentVolumeClaim",message="storageclass not found",event_namespace="test-ns4",event_object="pvc-opensearch-0",reason="ProvisioningFailed"} 500
kube_events_warning_total{controller="statefulset-controller",controller_instance="",kind="StatefulSet",message="Delete Pod error",event_namespace="postgres",event_object="pg-patroni-node2",reason="FailedDelete"} 1
kube_events_warning_total{controller="statefulset-controller",controller_instance="",kind="StatefulSet",message="StatefulSet is recreating failed Pod",event_namespace="postgres",event_object="pg-patroni-node1",reason="RecreatingFailedPod"} 29
# HELP kube_events_reporting_controller_normal_total Count of kubernetes events with type normal
# TYPE kube_events_reporting_controller_normal_total counter
kube_events_reporting_controller_normal_total{controller="",controller_instance="",kind="Certificate",event_namespace="test-ns1"} 4
kube_events_reporting_controller_normal_total{controller="",controller_instance="",kind="GrafanaDashboard",event_namespace="test-ns3"} 5
kube_events_reporting_controller_normal_total{controller="controllermanager",controller_instance="",kind="PodDisruptionBudget",event_namespace="ivsh-click"} 2
kube_events_reporting_controller_normal_total{controller="default-scheduler",controller_instance="default-scheduler-node-2",kind="Pod",event_namespace="mistral-vaze"} 6
kube_events_reporting_controller_normal_total{controller="default-scheduler",controller_instance="default-scheduler-node-2",kind="Pod",event_namespace="monitoring"} 1
kube_events_reporting_controller_normal_total{controller="deployment-controller",controller_instance="",kind="Deployment",event_namespace="test-ns3"} 4
kube_events_reporting_controller_normal_total{controller="job-controller",controller_instance="",kind="Job",event_namespace="default"} 4
kube_events_reporting_controller_normal_total{controller="kopf",controller_instance="dev",kind="MistralService",event_namespace="mistral-vaze"} 1
kube_events_reporting_controller_normal_total{controller="kubelet",controller_instance="node-1",kind="Pod",event_namespace="clickhouse-jyya"} 2
kube_events_reporting_controller_normal_total{controller="kubelet",controller_instance="node-1",kind="Pod",event_namespace="default"} 6
kube_events_reporting_controller_normal_total{controller="nginx-ingress-controller",controller_instance="",kind="ConfigMap",event_namespace="ingress-nginx"} 2
kube_events_reporting_controller_normal_total{controller="nginx-ingress-controller",controller_instance="",kind="Ingress",event_namespace="ams"} 14
kube_events_reporting_controller_normal_total{controller="rancher.io/local-path_local-path-provisioner-7bf8b8f4-25gds_ba26e854-edd7-47d4-b64b-df667ec1e279",controller_instance="",kind="PersistentVolumeClaim",event_namespace="dmsh-pg-test"} 12
# HELP kube_events_reporting_controller_warning_total Count of kubernetes events with type warning
# TYPE kube_events_reporting_controller_warning_total counter
kube_events_reporting_controller_warning_total{controller="",controller_instance="",kind="ClusterIssuer",event_namespace=""} 500
kube_events_reporting_controller_warning_total{controller="",controller_instance="",kind="GrafanaDashboard",event_namespace="test-ns1"} 1
kube_events_reporting_controller_warning_total{controller="controllermanager",controller_instance="",kind="PodDisruptionBudget",event_namespace="test-ns1"} 5
kube_events_reporting_controller_warning_total{controller="daemonset-controller",controller_instance="",kind="DaemonSet",event_namespace="test-ns2"} 156
kube_events_reporting_controller_warning_total{controller="endpoint-controller",controller_instance="",kind="Endpoints",event_namespace="test-ns3"} 13
kube_events_reporting_controller_warning_total{controller="kubelet",controller_instance="node-1",kind="Pod",event_namespace="test-ns3"} 4177
kube_events_reporting_controller_warning_total{controller="kubelet",controller_instance="node-2",kind="Pod",event_namespace="clickhouse-jyya"} 490
kube_events_reporting_controller_warning_total{controller="kubelet",controller_instance="node-3",kind="Pod",event_namespace="test-ns3"} 4156
kube_events_reporting_controller_warning_total{controller="persistentvolume-controller",controller_instance="",kind="PersistentVolumeClaim",event_namespace="test-ns4"} 1500
kube_events_reporting_controller_warning_total{controller="statefulset-controller",controller_instance="",kind="StatefulSet",event_namespace="postgres"} 57
```

<!-- markdownlint-enable line-length -->

### Event log example

This is an example of Event (API version events.k8s.io/v1):

<!-- markdownlint-disable line-length -->

```json
{
  "message": "Created new replication controller \"postgres-backup-daemon-1\" for version 1",
  "kind": "Event",
  "log": {
    "firstTimestamp": "2018-11-02T02:26:00Z",
    "reason": "DeploymentCreated",
    "metadata": {
      "name": "postgres-backup-daemon.15632d8844761c5d",
      "namespace": "pg96-nighttest",
      "creationTimestamp": "2018-11-02T02:26:00Z",
      "uid": "a38df948-de46-11e8-9fdb-fa163e5f2c4f",
      "selfLink": "/api/v1/namespaces/pg96-nighttest/events/postgres-backup-daemon.15632d8844761c5d",
      "resourceVersion": "71608432"
    },
    "apiVersion": "v1",
    "involvedObject": {
      "namespace": "pg96-nighttest",
      "name": "postgres-backup-daemon",
      "uid": "a388d4aa-de46-11e8-9fdb-fa163e5f2c4f",
      "apiVersion": "v1",
      "kind": "DeploymentConfig",
      "resourceVersion": "71608172"
    },
    "lastTimestamp": "2018-11-02T02:26:00Z",
    "count": 1,
    "source": {
      "component": "deploymentconfig-controller"
    },
    "type": "Normal"
  }
}
```

<!-- markdownlint-enable line-length -->

Formatted log output with default template of the example Event will be printed in stdout:

<!-- markdownlint-disable line-length -->

```text
{"time":"2018-11-02T02:26:00.999"}}","involvedObjectKind":"DeploymentConfig","involvedObjectNamespace":"pg96-nighttest","involvedObjectName":"postgres-backup-daemon","involvedObjectUid":"a388d4aa-de46-11e8-9fdb-fa163e5f2c4f","involvedObjectApiVersion":"v1","involvedObjectResourceVersion":"71608172","reason":"DeploymentCreated","type":"Normal","message":"Created new replication controller \"postgres-backup-daemon-1\" for version 1","kind":"Event"}
```

<!-- markdownlint-enable line-length -->

## Repository structure

* `./docs` - any documentation related to qubership-kube-events-reader
* `./pkg` - code of application
* `./pkg/aggregation` - mapping of events messages (related to events collected as metrics)
* `./pkg/controller` - kubernetes controller to watch Events
* `./pkg/filter` - logic of filtering events to exclude/include it to sink (stdout or metrics)
* `./pkg/format` - setting of events log format (related to events printed as logs)
* `./pkg/test` - testdata
* `./pkg/sink` - outputs of processed and filtered events
* `./pkg/utils` - general logic (logger, cli flags etc.)
* `./main.go` - application entrypoint

Files for build:

* `./Dockerfile` - to build Docker image

## How to start

If you are developer and need to make changes to this repository, please,
make sure that you read the information below.

### Build

Each time when you push your changes to repository CI pipeline is started.
Usually one of the latest step is application build and Docker image build.
So you do not need to run build job manually.

### Definition of done

After you made changes related to a task do next steps:

1. Check if there are any dependencies versions that can be upgraded in `go.mod`. Upgrade if it is possible.
2. Create tests if you modified behavior of application or fixed a bug (especially if it is not covered by tests).
3. Build qubership-kube-events-reader Docker image.
4. Check that all pipeline is succeded (linter, build, deploy & test jobs are passed).
5. Deploy qubership-kube-events-reader with [`qubership-logging-operator`](https://github.com/Netcracker/qubership-logging-operator/blob/main/docs/installation.md#cloud-events-reader)
   and [`qubership-monitoring-operator`](https://github.com/Netcracker/qubership-monitoring-operator/blob/main/docs/installation/components/exporters/cloud-events-exporter.md).
   Check that your feature works fine in possible cases.
6. Create merge request using merge request template. Name your MR `<TICKET-ID>: <SHORT-DESCRIPTION>`. Describe and
   explain your changes in MR. There you can add any information about the changes (how it was tested, details
   of aim of changes, examples and so on) to make it clear to the reviewers.

### Deploy

K8s events Reader is installed in Cloud as a part of:

1. Logging Service. Information about it`s installation
   described [here](https://github.com/Netcracker/qubership-logging-operator/blob/main/docs/installation.md#cloud-events-reader).
2. Platform Monitoring. Information about it`s installation
   described [here](https://github.com/Netcracker/qubership-monitoring-operator/blob/main/docs/installation/components/exporters/cloud-events-exporter.md).

#### Prerequisites

For `qubership-kube-events-reader` to work properly, in case of sending events as logs an instance of `FluentD` or `FluentBit`
should be installed in Kubernetes/Openshift.
In case of scraping events as metrics Monitoring components (VictoriaMetrics, Grafana etc.) have to be installed.

#### HWE and Limits

Events-reader is installed in Kubernetes/Openshift as a pod.

It requires:

* CPU: 100 millicores
* RAM: 128 MiB

#### Deploy with helm

To deploy qubership-kube-events-reader with qubership-logging-operator clone repository. Modify
[`charts/qubership-logging-operator/values.yaml`](https://github.com/Netcracker/qubership-logging-operator/blob/main/charts/qubership-logging-operator/values.yaml)
locally.

Set parameters:

```yaml
operatorImage: <logging-operator-image>
cloudEventsReader:
  install: true
  dockerImage: <qubership-kube-events-reader-image>
```

Run command to install logging-operator by Helm:

```bash
cd charts/logging-operator
helm install <any-release-name> . --namespace <namespace>
```

If you need to upgrade installation with the new parameters run:

```bash
helm upgrade <any-release-name> . --namespace <namespace>
```

To uninstall deployment run command:

```helm
helm uninstall <any-release-name> --namespace <namespace>
```

### How to debug

You can debug qubership-kube-events-reader locally with default or custom parameters in your IDE.
The only thing that you need to do before run/debug service locally you need to login and set context to the cloud.
qubership-kube-events-reader requires connection to cloud to watch for Kubernetes events.

### How to troubleshoot

There are no well-defined rules for troubleshooting, as each task is unique, but there are some tips that can do:

* See deployment parameters and cli flags
* See logs of the service
