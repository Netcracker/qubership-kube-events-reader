package main

import (
	"context"
	"testing"
	"time"
)

func TestShutdownRunsFinalizersWithShutdownContext(t *testing.T) {
	called := false

	err := Shutdown(context.Background(), time.Second, func(ctx context.Context) {
		called = true
		if ctx.Value(ContextKey) != ContextSd {
			t.Fatalf("expected shutdown context marker, got %v", ctx.Value(ContextKey))
		}
	})
	if err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
	if !called {
		t.Fatal("expected finalizer to be called")
	}
}

func TestShutdownReturnsTimeoutError(t *testing.T) {
	err := Shutdown(context.Background(), 10*time.Millisecond, func(ctx context.Context) {
		<-ctx.Done()
		time.Sleep(20 * time.Millisecond)
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestValidateFormatInput(t *testing.T) {
	if err := validateFormatInput("short-format"); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	longFormat := make([]byte, 1025)
	for i := range longFormat {
		longFormat[i] = 'a'
	}

	if err := validateFormatInput(string(longFormat)); err == nil {
		t.Fatal("expected oversized format validation error")
	}
}
