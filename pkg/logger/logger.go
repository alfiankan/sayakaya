package logger

import (
	"context"
	"github.com/sirupsen/logrus"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func Info(ctx context.Context, msg string, fields logrus.Fields) {
	entry := logrus.WithFields(fields)
	if traceID := GetTraceID(ctx); traceID != "" {
		entry = entry.WithField("trace_id", traceID)
	}
	entry.Info(msg)
}

func Error(ctx context.Context, msg string, err error, fields logrus.Fields) {
	entry := logrus.WithFields(fields)
	if traceID := GetTraceID(ctx); traceID != "" {
		entry = entry.WithField("trace_id", traceID)
	}
	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	entry.Error(msg)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	if traceID := GetTraceID(ctx); traceID != "" {
		entry = entry.WithField("trace_id", traceID)
	}
	entry.Infof(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	if traceID := GetTraceID(ctx); traceID != "" {
		entry = entry.WithField("trace_id", traceID)
	}
	entry.Errorf(format, args...)
}
