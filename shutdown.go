package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type ctxKey string

const (
	ContextSd         = "shutdownCtx"
	ContextKey ctxKey = "ctxKey"
)

type ReleaseFunc func(context.Context)

func Shutdown(ctx context.Context, timeout time.Duration, finalizers ...ReleaseFunc) error {
	slog.Info(fmt.Sprintf("trying to shut down gracefully, timeout %s", timeout.String()))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	done := make(chan struct{})
	go func(ctx context.Context) {
		for _, release := range finalizers {
			release(ctx)
		}
		done <- struct{}{}
	}(context.WithValue(ctx, ContextKey, ContextSd))
	select {
	case <-done:
		slog.Info("resources have been released")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to release resources in time")
	}
}
