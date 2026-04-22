package utils

import (
	"log/slog"
	"testing"
	"time"
)

func TestReplaceAttrsFormatsTopLevelTime(t *testing.T) {
	ts := time.Date(2026, time.April, 17, 10, 11, 12, 123000000, time.UTC)

	attr := ReplaceAttrs(nil, slog.Any(slog.TimeKey, ts))

	if attr.Key != slog.TimeKey {
		t.Fatalf("expected time key, got %q", attr.Key)
	}
	if attr.Value.String() != "2026-04-17T10:11:12.123" {
		t.Fatalf("unexpected formatted time: %q", attr.Value.String())
	}
}

func TestReplaceAttrsFormatsSourceFile(t *testing.T) {
	attr := ReplaceAttrs(nil, slog.Any(slog.SourceKey, &slog.Source{File: "/tmp/example/file.go", Line: 42}))

	if attr.Key != slog.SourceKey {
		t.Fatalf("expected source key, got %q", attr.Key)
	}
	if attr.Value.String() != "file.go:42" {
		t.Fatalf("unexpected formatted source: %q", attr.Value.String())
	}
}

func TestReplaceAttrsLeavesGroupedTimeUntouched(t *testing.T) {
	original := slog.Any(slog.TimeKey, time.Date(2026, time.April, 17, 10, 11, 12, 0, time.UTC))

	attr := ReplaceAttrs([]string{"group"}, original)

	if attr.Key != original.Key {
		t.Fatalf("expected original key to be preserved, got %q", attr.Key)
	}
	if attr.Value.Any().(time.Time) != original.Value.Any().(time.Time) {
		t.Fatal("expected grouped time attribute to remain unchanged")
	}
}
