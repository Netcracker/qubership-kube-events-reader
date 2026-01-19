package main

import (
	"context"
	"testing"
	"time"
)

func TestShutdown_Success(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	called := false
	finalizer := func(ctx context.Context) {
		called = true
		time.Sleep(100 * time.Millisecond) // Short delay
	}

	err := Shutdown(ctx, timeout, finalizer)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected finalizer to be called")
	}
}

func TestShutdown_Timeout(t *testing.T) {
	ctx := context.Background()
	timeout := 100 * time.Millisecond

	finalizer := func(ctx context.Context) {
		time.Sleep(200 * time.Millisecond) // Longer than timeout
	}

	err := Shutdown(ctx, timeout, finalizer)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestShutdown_MultipleFinalizers(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	callCount := 0
	finalizer1 := func(ctx context.Context) {
		callCount++
	}
	finalizer2 := func(ctx context.Context) {
		callCount++
	}

	err := Shutdown(ctx, timeout, finalizer1, finalizer2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 finalizers called, got %d", callCount)
	}
}
