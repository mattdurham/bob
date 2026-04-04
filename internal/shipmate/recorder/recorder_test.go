package recorder

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTestRecorder(t *testing.T) (*Recorder, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	rec, err := New(exp)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		rec.Shutdown(context.Background()) //nolint:errcheck
	})
	return rec, exp
}

func TestRecordCreatesSpan(t *testing.T) {
	rec, exp := newTestRecorder(t)

	args := RecordArgs{
		Name:  "test-span",
		Agent: "coder-1",
		Text:  "hello",
		Attributes: map[string]string{
			"k": "v",
		},
	}
	if err := rec.Record(context.Background(), args); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := rec.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	if span.Name != "test-span" {
		t.Errorf("span name: got %q, want %q", span.Name, "test-span")
	}

	attrs := spanAttrsMap(span)
	if got := attrs["memory.agent"]; got != "coder-1" {
		t.Errorf("shipmate.agent: got %q, want %q", got, "coder-1")
	}
	if got := attrs["memory.text"]; got != "hello" {
		t.Errorf("shipmate.text: got %q, want %q", got, "hello")
	}
	if got := attrs["k"]; got != "v" {
		t.Errorf("k: got %q, want %q", got, "v")
	}

	// service.name is set on the Resource, not individual span attributes.
	resAttrs := resourceAttrsMap(span)
	if got := resAttrs["service.name"]; got != "shipmate" {
		t.Errorf("service.name on resource: got %q, want %q", got, "shipmate")
	}
}

func TestRecordStampsSessionID(t *testing.T) {
	rec, exp := newTestRecorder(t)

	args := RecordArgs{
		Name:      "session-span",
		Agent:     "coder-2",
		Text:      "with session",
		SessionID: "ses-42",
	}
	if err := rec.Record(context.Background(), args); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := rec.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := spanAttrsMap(spans[0])
	if got := attrs["session.id"]; got != "ses-42" {
		t.Errorf("session.id: got %q, want %q", got, "ses-42")
	}
}

func TestRecordNoSessionID(t *testing.T) {
	rec, exp := newTestRecorder(t)

	args := RecordArgs{
		Name:      "no-session-span",
		Agent:     "coder-2",
		Text:      "no session",
		SessionID: "",
	}
	if err := rec.Record(context.Background(), args); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := rec.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := spanAttrsMap(spans[0])
	if _, found := attrs["session.id"]; found {
		t.Errorf("expected no session.id attribute when SessionID is empty, but found it")
	}
}

func TestRecordCustomAttributes(t *testing.T) {
	rec, exp := newTestRecorder(t)

	args := RecordArgs{
		Name:  "attr-span",
		Agent: "coder-2",
		Text:  "attrs",
		Attributes: map[string]string{
			"repo":  "bob",
			"task":  "3",
			"phase": "execute",
		},
	}
	if err := rec.Record(context.Background(), args); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := rec.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := spanAttrsMap(spans[0])

	for k, want := range map[string]string{"repo": "bob", "task": "3", "phase": "execute"} {
		if got := attrs[k]; got != want {
			t.Errorf("%s: got %q, want %q", k, got, want)
		}
	}
}

// spanAttrsMap converts span attributes to a plain string map for easy lookup.
func spanAttrsMap(span tracetest.SpanStub) map[string]string {
	m := make(map[string]string, len(span.Attributes))
	for _, kv := range span.Attributes {
		m[string(kv.Key)] = kv.Value.AsString()
	}
	return m
}

// resourceAttrsMap converts a span's resource attributes to a plain string map.
func resourceAttrsMap(span tracetest.SpanStub) map[string]string {
	if span.Resource == nil {
		return nil
	}
	attrs := span.Resource.Attributes()
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.AsString()
	}
	return m
}

// Compile-time check that Recorder uses the SpanExporter interface.
var _ sdktrace.SpanExporter = (*tracetest.InMemoryExporter)(nil)
