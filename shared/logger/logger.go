// Package logger provides a single structured JSON logger (via log/slog)
// used by every StreamPass module. print()/fmt.Println() must never be used
// for application logging — this package is the only sanctioned entry point.
package logger

import (
	"context"
	"log/slog"
	"os"

	apperrors "streampass/shared/errors"
)

// Logger wraps slog.Logger and pins module/operation tagging so call sites
// don't repeat boilerplate attributes on every call.
type Logger struct {
	base *slog.Logger
}

// New builds a Logger that writes structured JSON to stdout at the given
// level ("debug", "info", "warn", "error"; defaults to "info").
func New(module string, level string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return &Logger{base: slog.New(handler).With(slog.String("module", module))}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// With returns a child Logger tagged with the given operation, so every
// subsequent log line from a single use-case call carries it automatically.
func (l *Logger) With(operation string) *Logger {
	return &Logger{base: l.base.With(slog.String("operation", operation))}
}

// Info logs an informational structured event.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.base.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Debug logs a debug-level structured event.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.base.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Warn logs a warning-level structured event.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.base.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error logs an *AppError (or any error) with code, message and cause
// extracted into their own structured fields, per spec section "Ошибки":
// every error entry must carry code, message and cause.
func (l *Logger) Error(ctx context.Context, err error, attrs ...slog.Attr) {
	code := apperrors.CodeOf(err)
	base := []slog.Attr{
		slog.String("error_code", string(code)),
		slog.String("error", err.Error()),
	}
	l.base.LogAttrs(ctx, slog.LevelError, "operation failed", append(base, attrs...)...)
}
