// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logger contains util functions to get/set context to	q log fields
package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logger that is configured. It is to be used by arbitrary packages when no
// logger instance has been passed.
var (
	Logger *zap.Logger
)

type ctxValueType string

const (
	loggerKey ctxValueType = "x-logger"

	TraceSampledJsonKey        = "traceSampled"
	TraceIDJsonKey      string = "traceId"
	SpanIDJsonKey       string = "spanId"
)

// FromContext returns a logger from given context. If context has no
// logger, it uses application default logger initialized in config
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}

	return Logger
}

// AddTraceContext adds trace details to the log context.
func AddTraceContext(ctx context.Context) context.Context {
	// Get SpanContext from context
	spanCtx := trace.SpanContextFromContext(ctx)

	return AddSpanCtx(ctx, spanCtx)
}

// AddSpanCtx adds trace details from span context to the given context.
func AddSpanCtx(ctx context.Context, spanCtx trace.SpanContext) context.Context {
	// Extract Trace and Span IDs
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()

	// RFC requires traceId tag in logs. Maintain compatibility
	traceFields := []zap.Field{
		zap.String(TraceIDJsonKey, traceID),
		zap.String(SpanIDJsonKey, spanID),
		zap.Bool(TraceSampledJsonKey, spanCtx.IsSampled()),
	}

	// Add fields to existing log instance
	log := FromContext(ctx).With(traceFields...)

	// Add log instance to context for future use
	return ToContext(ctx, log)
}

// ToContext adds given logger to given context
func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
