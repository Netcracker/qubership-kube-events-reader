package sink

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/aggregation"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	versionCollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
)

var (
	versionGauge   = versionCollector.NewCollector("kube_events_exporter")
	SummaryCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_total",
		Help: "Count of kubernetes events",
	},
		[]string{"kind", "event_namespace", "type"},
	)
	NormalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_normal_total",
		Help: "Count of kubernetes events with type normal aggregated by message",
	},
		[]string{"kind", "event_object", "event_namespace", "reason", "controller", "controller_instance", "message"},
	)
	WarningCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_warning_total",
		Help: "Count of kubernetes events with type warning aggregated by message",
	},
		[]string{"kind", "event_object", "event_namespace", "reason", "controller", "controller_instance", "message"},
	)
	ReportingControllerNormalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_reporting_controller_normal_total",
		Help: "Count of kubernetes events with type normal",
	},
		[]string{"controller", "controller_instance", "kind", "event_namespace"},
	)
	ReportingControllerWarningCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_reporting_controller_warning_total",
		Help: "Count of kubernetes events with type warning",
	},
		[]string{"controller", "controller_instance", "kind", "event_namespace"},
	)
)

type PrometheusMetricsSink struct {
	*Sink
}

func InitMetricsSink(ctx context.Context, port string, metricsPath string, filters *filter.Sink, startHttpEndpoint func(context.Context, string)) (*PrometheusMetricsSink, error) {
	return InitMetricsSinkWithRegistry(ctx, port, metricsPath, filters, startHttpEndpoint, nil)
}

func InitMetricsSinkWithRegistry(ctx context.Context, port string, metricsPath string, filters *filter.Sink, startHttpEndpoint func(context.Context, string), registry prometheus.Registerer) (*PrometheusMetricsSink, error) {
	if !utils.IsPortValid(port) {
		return nil, fmt.Errorf("port is not valid for metrics endpoint. Given value: %v", port)
	}
	sink := initializeSinkWithFilters(filters)
	aggregation.InitAggregations()
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}
	if startHttpEndpoint == nil {
		registerMetricsToRegistry(registry)
		if registry == prometheus.DefaultRegisterer {
			startMetricsEndpoint(ctx, port, metricsPath)
		}
	} else {
		registerMetricsToRegistry(registry)
		startHttpEndpoint(ctx, port)
	}
	return &PrometheusMetricsSink{Sink: sink}, nil
}

func startMetricsEndpoint(ctx context.Context, port string, path string) {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.Handler())
	srv := http.Server{
		Addr:         fmt.Sprintf(":%v", port),
		Handler:      mux,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 30,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		exit := srv.ListenAndServe()
		if !errors.Is(exit, http.ErrServerClosed) {
			slog.Error(fmt.Sprintf("failed to start HTTP server. Error: %s", exit))
		}
	}()
}

func (ms *PrometheusMetricsSink) Release(eventObj *corev1.Event) error {
	if !ms.IsEventAllowed(eventObj) {
		return nil
	}

	// Sanitize label values only for required fields that should always have values
	kind := eventObj.InvolvedObject.Kind
	if kind == "" {
		kind = "unknown"
	}
	namespace := eventObj.InvolvedObject.Namespace
	if namespace == "" {
		namespace = "unknown"
	}
	eventType := eventObj.Type
	if eventType == "" {
		eventType = "unknown"
	}
	name := eventObj.InvolvedObject.Name
	if name == "" {
		name = "unknown"
	}
	reason := eventObj.Reason
	if reason == "" {
		reason = "unknown"
	}
	// Keep optional fields as empty strings if legitimately empty
	reportingController := eventObj.ReportingController
	reportingInstance := eventObj.ReportingInstance
	message := aggregation.GetCommonMessage(eventObj.InvolvedObject.Kind, eventObj.Reason, eventObj.Message)

	SummaryCounter.WithLabelValues(kind, namespace, eventType).Inc()
	if strings.EqualFold(eventObj.Type, corev1.EventTypeNormal) {
		ReportingControllerNormalCounter.WithLabelValues(reportingController, reportingInstance, kind, namespace).Inc()
		NormalCounter.WithLabelValues(kind, name, namespace, reason, reportingController, reportingInstance, message).Inc()
	} else {
		ReportingControllerWarningCounter.WithLabelValues(reportingController, reportingInstance, kind, namespace).Inc()
		WarningCounter.WithLabelValues(kind, name, namespace, reason, reportingController, reportingInstance, message).Inc()
	}
	return nil
}

func registerMetricsToRegistry(registry prometheus.Registerer) {
	registry.MustRegister(versionGauge, SummaryCounter, NormalCounter, WarningCounter, ReportingControllerNormalCounter, ReportingControllerWarningCounter)
}

func UnregisterMetrics() {
	prometheus.Unregister(versionGauge)
	prometheus.Unregister(SummaryCounter)
	prometheus.Unregister(NormalCounter)
	prometheus.Unregister(WarningCounter)
	prometheus.Unregister(ReportingControllerNormalCounter)
	prometheus.Unregister(ReportingControllerWarningCounter)
	SummaryCounter.Reset()
	NormalCounter.Reset()
	WarningCounter.Reset()
	ReportingControllerNormalCounter.Reset()
	ReportingControllerWarningCounter.Reset()
}
