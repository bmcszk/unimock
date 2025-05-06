package logger

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Logger wraps slog.Logger with additional methods
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new instance of Logger
func NewLogger() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}

// LogRequest logs HTTP request details
func (l *Logger) LogRequest(r *http.Request, statusCode int, duration time.Duration) {
	l.Info("http request",
		"method", r.Method,
		"path", r.URL.Path,
		"content_type", r.Header.Get("Content-Type"),
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)
}

// LogStorage logs storage operations
func (l *Logger) LogStorage(operation string, ids []string, success bool) {
	l.Info("storage operation",
		"operation", operation,
		"ids", ids,
		"success", success,
	)
}

// LogError logs error details
func (l *Logger) LogError(msg string, err error, attrs ...any) {
	args := make([]any, 0, len(attrs)+2)
	args = append(args, "error", err)
	args = append(args, attrs...)
	l.Error(msg, args...)
}
