// Package recorder creates synthetic already-ended OTEL spans and exports them.
// It is used by the shipmate MCP server to allow agents to annotate traces.
package recorder

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// RecordArgs holds the arguments from the shipmate_record MCP tool call.
type RecordArgs struct {
	// Name is the span name.
	Name string
	// Agent is the agent identity, e.g. "coder-1".
	Agent string
	// Text is a free-form description of what the agent did.
	Text string
	// SessionID is the current Claude Code session.id. If empty, the attribute
	// is omitted entirely from the synthetic span.
	SessionID string
	// Attributes holds optional key-value pairs to add to the span.
	Attributes map[string]string
}

// Recorder creates synthetic already-ended spans and exports them.
type Recorder struct {
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
}

// New creates a Recorder that exports spans via the given SpanExporter.
// It sets service.name="shipmate" on the TracerProvider resource and uses
// sdktrace.WithBatcher for non-blocking, buffered export. Spans are queued
// in memory and exported in background batches. shipmate_record returns
// immediately regardless of upstream availability.
func New(exp sdktrace.SpanExporter) (*Recorder, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("shipmate"),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	return &Recorder{
		tracer: tp.Tracer("shipmate"),
		tp:     tp,
	}, nil
}

// Record creates and immediately ends a synthetic span with the given args.
// Export is non-blocking — spans are queued and flushed in background batches.
// session.id is omitted entirely when args.SessionID is empty.
func (r *Recorder) Record(ctx context.Context, args RecordArgs) error {
	attrs := []attribute.KeyValue{
		attribute.String("shipmate.agent", args.Agent),
		attribute.String("shipmate.text", args.Text),
	}
	if args.SessionID != "" {
		attrs = append(attrs, attribute.String("session.id", args.SessionID))
	}
	for k, v := range args.Attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	_, span := r.tracer.Start(ctx, args.Name,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attrs...),
	)
	span.End()
	return nil
}

// ForceFlush immediately exports all queued spans. Useful in tests.
func (r *Recorder) ForceFlush(ctx context.Context) error {
	return r.tp.ForceFlush(ctx)
}

// Shutdown flushes and closes the underlying TracerProvider.
func (r *Recorder) Shutdown(ctx context.Context) error {
	return r.tp.Shutdown(ctx)
}
