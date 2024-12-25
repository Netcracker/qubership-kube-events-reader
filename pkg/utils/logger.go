package utils

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"
)

func ReplaceAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		actualTime := a.Value.Any().(time.Time)
		timeValue := actualTime.Format("2006-01-02T15:04:05.999")
		return slog.Attr{Key: slog.TimeKey, Value: slog.StringValue(timeValue)}
	}
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
		return slog.Attr{
			Key:   slog.SourceKey,
			Value: slog.StringValue(fmt.Sprintf("%s:%v", filepath.Base(source.File), source.Line)),
		}
	}
	return a
}
