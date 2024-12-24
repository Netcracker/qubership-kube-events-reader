package utils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"regexp"
	"time"
)

func StartHealthEndpoint(ctx context.Context, pprofEnabled bool, port string) (*http.Server, error) {
	if !IsPortValid(port) {
		return nil, fmt.Errorf("port is not valid for metrics endpoint. Given value: %v", port)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)
	if pprofEnabled {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/{action}", pprof.Index)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}
	srv := http.Server{
		Addr:         net.JoinHostPort("", port),
		Handler:      mux,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
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
	return &srv, nil
}

func healthCheck(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Header().Add("ContentType", "application/json")
		_, _ = responseWriter.Write([]byte("ok"))
	}
}

func IsPortValid(port string) bool {
	portRegexp := regexp.MustCompile(`\d{4,6}`)
	return portRegexp.MatchString(port)
}
