package utils

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
)

// Logger is the global logger instance
var Logger *log.Logger

// ContextKey is used for context keys
type ContextKey string

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

func init() {
	Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
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