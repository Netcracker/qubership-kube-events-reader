package utils

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestStartHealthEndpointRejectsInvalidPort(t *testing.T) {
	srv, err := StartHealthEndpoint(context.Background(), false, "abc")
	if err == nil {
		t.Fatal("expected invalid port error")
	}
	if srv != nil {
		t.Fatal("expected nil server for invalid port")
	}
}

func TestStartHealthEndpointRegistersHandlers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	l.Close()

	srv, err := StartHealthEndpoint(ctx, true, port)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	srv.Handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "ok" {
		t.Fatalf("expected health body to be ok, got %q", recorder.Body.String())
	}

	pprofReq := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	pprofRecorder := httptest.NewRecorder()
	srv.Handler.ServeHTTP(pprofRecorder, pprofReq)
	if pprofRecorder.Code != http.StatusOK {
		t.Fatalf("expected pprof endpoint to be registered, got %d", pprofRecorder.Code)
	}
}

func TestHealthCheckIgnoresNonGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	healthCheck(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected default recorder status for ignored method, got %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body for ignored method, got %q", recorder.Body.String())
	}
}

func TestIsPortValid(t *testing.T) {
	testCases := []struct {
		port  string
		valid bool
	}{
		{port: "0", valid: false},
		{port: "1", valid: true},
		{port: "999", valid: true},
		{port: "8080", valid: true},
		{port: "65535", valid: true},
		{port: "65536", valid: false},
		{port: "123456", valid: false},
		{port: "1234567", valid: false},
		{port: "port8080", valid: false},
		{port: "abc", valid: false},
		{port: "", valid: false},
	}

	for _, testCase := range testCases {
		if got := IsPortValid(testCase.port); got != testCase.valid {
			t.Fatalf("expected validity for %q to be %v, got %v", testCase.port, testCase.valid, got)
		}
	}
}
