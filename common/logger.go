// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type contextKey string

const RequestIDKey contextKey = "requestid"

// HourlyLogWriter is a thread-safe writer that writes logs to stdout and automatically
// rotates logs to hourly files in the 'logs/' folder.
// This rotation prevents single log files from growing too large in production.
type HourlyLogWriter struct {
	mu       sync.Mutex
	file     *os.File
	lastHour int
	lastDay  int
}

func (w *HourlyLogWriter) Write(p []byte) (n int, err error) {
	// 1. Always write to stdout so log collectors like Docker/Kubernetes can capture them
	os.Stdout.Write(p)

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	hour := now.Hour()
	day := now.Day()

	// 2. Check if file is not initialized or hour/day has changed for rotation
	if w.file == nil || hour != w.lastHour || day != w.lastDay {
		if w.file != nil {
			_ = w.file.Close()
		}

		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			_ = os.Mkdir("logs", 0770)
		}

		fileName := fmt.Sprintf("logs/log_%s.log", now.Format("01-02-2006_15"))
		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
		w.file = f
		w.lastHour = hour
		w.lastDay = day
	}

	return w.file.Write(p)
}

// RequestIDHook is a logrus Hook that automatically filters context
// and appends the "request_id" field to the JSON output if it exists in the context.
type RequestIDHook struct{}

func (h *RequestIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *RequestIDHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		// Look up with custom contextKey type
		if reqID, ok := entry.Context.Value(RequestIDKey).(string); ok && reqID != "" {
			entry.Data["request_id"] = reqID
			return nil
		}
		// Fallback to plain string type
		if reqID, ok := entry.Context.Value("requestid").(string); ok && reqID != "" {
			entry.Data["request_id"] = reqID
		}
	}
	return nil
}

// Log is the global singleton logger
var Log *logrus.Logger

func init() {
	Log = logrus.New()
	Log.SetLevel(logrus.InfoLevel)
	Log.SetReportCaller(true) // Attaches file name and line number to logs
	Log.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	})
	Log.SetOutput(&HourlyLogWriter{})
	Log.AddHook(&RequestIDHook{})
}

// Logger returns a logrus Entry carrying context (for request ID tracing)
// and tags the log entry with a specific "scope" (such as a Controller or Service name).
//
// Controller/Service Usage Example:
//
//	func (c *ProductController) Create(ctx context.Context) {
//	    common.Logger(ctx, "ProductController").Info("Starting product creation process")
//	}
//
// Example JSON Output:
//
//	{
//	  "@timestamp": "2026-05-31T14:15:30Z",
//	  "level": "info",
//	  "scope": "ProductController",
//	  "request_id": "7ac15b82-f8c3-4d40-bb53-fa91c8cf7de1",
//	  "file": "product_controller.go:37",
//	  "message": "Starting product creation process"
//	}
func Logger(ctx context.Context, scopes ...string) *logrus.Entry {
	entry := Log.WithContext(ctx)
	if len(scopes) > 0 {
		entry = entry.WithField("scope", scopes[0])
	}
	return entry
}

// NewLogger is kept for backward compatibility with external templates,
// mapping instance redirections to the global singleton Log to prevent memory leaks.
func NewLogger() *logrus.Logger {
	return Log
}
