package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsPortValid(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{"valid 4-digit port", "8080", true},
		{"valid 5-digit port", "99999", true},
		{"valid 6-digit port", "65535", true},
		{"valid port with colon", ":9999", true},
		{"invalid short port", "123", false},
		{"invalid long port", "1234567", false},
		{"invalid non-numeric", "abcd", false},
		{"invalid empty", "", false},
		{"invalid with letters", "8080a", false},
		{"invalid double colon", "::9999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPortValid(tt.port)
			if result != tt.expected {
				t.Errorf("IsPortValid(%q) = %v, want %v", tt.port, result, tt.expected)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthCheck(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	expected := "ok"
	if string(body) != expected {
		t.Errorf("Expected body %q, got %q", expected, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestStartHealthEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := "8080" // Use a valid port
	srv, err := StartHealthEndpoint(ctx, false, port)
	if err != nil {
		t.Fatalf("StartHealthEndpoint failed: %v", err)
	}
	defer srv.Close()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get("http://localhost:" + port + "/health")
	if err != nil {
		t.Fatalf("Failed to make request to health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	expected := "ok"
	if string(body) != expected {
		t.Errorf("Expected body %q, got %q", expected, string(body))
	}
}

func TestStartHealthEndpoint_InvalidPort(t *testing.T) {
	ctx := context.Background()
	_, err := StartHealthEndpoint(ctx, false, "invalid")
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}
