package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/controller"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/sink"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/utils"
	"github.com/go-logr/logr"
	_ "go.uber.org/automaxprocs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	logsType    = "logs"
	metricsType = "metrics"
)

var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: utils.ReplaceAttrs, AddSource: true}))

func init() {
	slog.SetDefault(Logger)
	// Setting Json log handler to klog that is used by client-go
	klog.SetLogger(logr.FromSlogHandler(Logger.Handler()))
}

func main() {
	var namespaceFlags utils.NamespaceFlagsType
	flag.Var(&namespaceFlags, "namespace", "Namespace to watch for events. The parameter can be used multiple times. If parameter is not set events of all namespaces will be watched")
	var outputs utils.SinksFlagsType
	flag.Var(&outputs, "output", "Outputs for events. The parameter can be used multiple times. The parameter has two available values: metrics or/and logs.")
	workers := flag.Int("workers", 2, "Workers number for controller")
	printFormat := flag.String("format", "", "Format to print Event. It should be valid Golang template of `text/template` package")
	metricsPort := flag.String("metricsPort", "9999", "Port to expose Prometheus metrics on")
	metricsPath := flag.String("metricsPath", "/metrics", "HTTP path to scrape for Prometheus metrics")
	filterFile := flag.String("filtersPath", "", "Absolute path to file with filter events configuration")
	pprofEnabled := flag.Bool("pprofEnable", true, "Enable pprof")
	healthServePort := flag.String("pprofAddr", "8080", "Port to health and pprof endpoint")
	flag.Parse()

	slog.Info("starting K8s Events Reader...")

	if len(outputs) < 1 {
		slog.Warn("output for events is not set. Output in logs is used by default")
		if err := outputs.Set(logsType); err != nil {
			slog.Error("could not set output")
			os.Exit(1)
		}
	}

	filters, err := filter.ParseFiltersConfiguration(*filterFile)
	if err != nil {
		slog.Error("could not parse filter events configuration. See `filtersPath` parameters and content of the file")
		os.Exit(1)
	}

	cfg := ctrl.GetConfigOrDie()
	cfg.ContentType = "application/vnd.kubernetes.protobuf"
	cfg.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(5, 10)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.Error("error building kubernetes client set", "error", err)
		os.Exit(1)
	}

	srvBaseCtx := signals.SetupSignalHandler()
	var sinks []sink.ISink
	if slices.Contains(outputs, logsType) {
		stdoutSink, err := sink.InitStdoutSink(*printFormat, filters.GetSinkFiltersByName(logsType))
		if err != nil {
			slog.Error("error occurred during initialization of logs output", "error", err)
			os.Exit(1)
		}
		sinks = append(sinks, stdoutSink)
		slog.Info("sink initialized successfully", "sink", "stdout")
	}
	if slices.Contains(outputs, metricsType) {
		metricsSink, err := sink.InitMetricsSink(srvBaseCtx, *metricsPort, *metricsPath, filters.GetSinkFiltersByName(metricsType), nil)
		if err != nil {
			slog.Error("error occurred during initialization of metrics output", "error", err)
			os.Exit(1)
		}
		sinks = append(sinks, metricsSink)
		slog.Info("sink initialized successfully", "sink", "metrics")
	}
	filters = nil

	var controllers []*controller.EventController
	observedNamespaces := strings.Split(namespaceFlags.String(), ",")
	if len(observedNamespaces) == 1 && observedNamespaces[0] == "" {
		controllers = append(controllers, controller.NewClusterEventController(kubeClient, controller.NewListerWatcherFunc(), sinks))
	} else {
		controllers = controller.NewNamespacedEventControllers(kubeClient, observedNamespaces, controller.NewListerWatcherFunc(), sinks)
	}
	stop := make(chan struct{})
	defer close(stop)
	for _, c := range controllers {
		go c.Run(*workers, stop)
	}

	srv, err := utils.StartHealthEndpoint(srvBaseCtx, *pprofEnabled, *healthServePort)
	if err != nil {
		slog.Error("could not start health endpoint", "error", err)
	}

	<-srvBaseCtx.Done()
	slog.Info("stopping application")

	if err = Shutdown(srvBaseCtx, 30*time.Second,
		func(ctx context.Context) {
			if err = srv.Shutdown(ctx); err != nil {
				slog.Error(fmt.Sprintf("failed to shut down HTTP server gracefully in time. Error: %s", err))
				slog.Info("force closing http server", "error", srv.Close())
			}
			slog.Info("http server is shut down")
		},
	); err != nil {
		slog.Error(fmt.Sprintf("failed to shutdown gracefully. Error: %s", err))
	}
}
