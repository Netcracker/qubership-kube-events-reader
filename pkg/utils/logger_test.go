package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestReplaceAttrs_TimeKey(t *testing.T) {
	// Test time key replacement
	now := time.Date(2023, 1, 1, 12, 30, 45, 123456789, time.UTC)
	attr := slog.Attr{Key: slog.TimeKey, Value: slog.AnyValue(now)}

	result := ReplaceAttrs([]string{}, attr)

	expectedTimeStr := "2023-01-01T12:30:45.123"
	if result.Key != slog.TimeKey {
		t.Errorf("Expected key %s, got %s", slog.TimeKey, result.Key)
	}
	if result.Value.String() != expectedTimeStr {
		t.Errorf("Expected time string %q, got %q", expectedTimeStr, result.Value.String())
	}
}

func TestReplaceAttrs_SourceKey(t *testing.T) {
	// Test source key replacement
	source := &slog.Source{
		File: "/path/to/file.go",
		Line: 42,
	}
	attr := slog.Attr{Key: slog.SourceKey, Value: slog.AnyValue(source)}

	result := ReplaceAttrs([]string{}, attr)

	if result.Key != slog.SourceKey {
		t.Errorf("Expected key %s, got %s", slog.SourceKey, result.Key)
	}

	expected := "file.go:42"
	if result.Value.String() != expected {
		t.Errorf("Expected source string %q, got %q", expected, result.Value.String())
	}
}

func TestReplaceAttrs_OtherKey(t *testing.T) {
	// Test other keys are returned unchanged
	attr := slog.Attr{Key: "other", Value: slog.StringValue("value")}

	result := ReplaceAttrs([]string{}, attr)

	if result.Key != attr.Key || result.Value.String() != attr.Value.String() {
		t.Errorf("Expected attr to be unchanged, got %+v", result)
	}
}

func TestReplaceAttrs_WithGroups(t *testing.T) {
	// Test that time key is not modified when groups are present
	now := time.Date(2023, 1, 1, 12, 30, 45, 123456789, time.UTC)
	attr := slog.Attr{Key: slog.TimeKey, Value: slog.AnyValue(now)}

	result := ReplaceAttrs([]string{"group"}, attr)

	// Should return the original attr unchanged
	if result.Key != attr.Key || result.Value.Any() != attr.Value.Any() {
		t.Errorf("Expected attr to be unchanged when groups present, got %+v", result)
	}
}

func TestLoggerIntegration(t *testing.T) {
	// Test the logger with ReplaceAttrs in action
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: ReplaceAttrs,
		AddSource:   true,
	})
	logger := slog.New(handler)

	// Log a message
	logger.Info("test message")

	output := buf.String()

	// Check that time is formatted correctly
	if !strings.Contains(output, `"time":"20`) {
		t.Errorf("Expected formatted time in output, got: %s", output)
	}

	// Check that source file is basename only
	if !strings.Contains(output, `"source":"logger_test.go:`) {
		t.Errorf("Expected basename in source file, got: %s", output)
	}
}
