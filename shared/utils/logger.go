package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// ContextKey is used for context keys
type ContextKey string

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// ServiceNameKey is the context key for service name
	ServiceNameKey ContextKey = "service_name"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"trace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	ServiceName string                 `json:"service_name,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// StructuredLogger provides structured logging functionality
type StructuredLogger struct {
	output      io.Writer
	level       LogLevel
	serviceName string
}

// Logger is the global logger instance
var Logger *StructuredLogger

func init() {
	Logger = NewStructuredLogger(os.Stdout, LogLevelInfo, "shopsphere")
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(output io.Writer, level LogLevel, serviceName string) *StructuredLogger {
	return &StructuredLogger{
		output:      output,
		level:       level,
		serviceName: serviceName,
	}
}

// SetLevel sets the log level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetServiceName sets the service name
func (l *StructuredLogger) SetServiceName(serviceName string) {
	l.serviceName = serviceName
}

// shouldLog determines if a message should be logged based on level
func (l *StructuredLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}
	return levels[level] >= levels[l.level]
}

// log writes a log entry
func (l *StructuredLogger) log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}, err error) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     message,
		TraceID:     GetTraceID(ctx),
		UserID:      GetUserID(ctx),
		ServiceName: GetServiceName(ctx),
		RequestID:   GetRequestID(ctx),
		Fields:      fields,
	}

	if entry.ServiceName == "" {
		entry.ServiceName = l.serviceName
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Get caller information
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.File = file
		entry.Line = line
		if fn := runtime.FuncForPC(pc); fn != nil {
			entry.Function = fn.Name()
		}
	}

	// Marshal and write the log entry
	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ctx, LogLevelDebug, message, f, nil)
}

// Info logs an info message
func (l *StructuredLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ctx, LogLevelInfo, message, f, nil)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ctx, LogLevelWarn, message, f, nil)
}

// Error logs an error message
func (l *StructuredLogger) Error(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ctx, LogLevelError, message, f, err)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ctx, LogLevelFatal, message, f, err)
	os.Exit(1)
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithServiceName adds a service name to the context
func WithServiceName(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, ServiceNameKey, serviceName)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		return traceID.(string)
	}
	return ""
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID := ctx.Value(UserIDKey); userID != nil {
		return userID.(string)
	}
	return ""
}

// GetServiceName retrieves the service name from context
func GetServiceName(ctx context.Context) string {
	if serviceName := ctx.Value(ServiceNameKey); serviceName != nil {
		return serviceName.(string)
	}
	return ""
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		return requestID.(string)
	}
	return ""
}

// LogMiddleware creates a logging middleware for HTTP requests
func LogMiddleware(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Generate request ID and trace ID
			requestID := uuid.New().String()
			ctx := WithRequestID(r.Context(), requestID)
			ctx = WithTraceID(ctx)
			ctx = WithServiceName(ctx, serviceName)
			
			// Add trace ID to response headers
			w.Header().Set("X-Trace-ID", GetTraceID(ctx))
			w.Header().Set("X-Request-ID", requestID)
			
			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
			
			// Log request
			Logger.Info(ctx, "HTTP request started", map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"user_agent": r.UserAgent(),
				"remote_ip":  getClientIP(r),
			})
			
			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))
			
			// Log response
			duration := time.Since(start)
			Logger.Info(ctx, "HTTP request completed", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status_code": wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
			})
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}