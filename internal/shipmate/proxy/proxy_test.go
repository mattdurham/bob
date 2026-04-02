package proxy

import (
	"context"
	"testing"

	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// mockForwarder records the last Export call for assertions.
type mockForwarder struct {
	calls []*collectortrace.ExportTraceServiceRequest
	resp  *collectortrace.ExportTraceServiceResponse
	err   error
}

func (m *mockForwarder) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	m.calls = append(m.calls, req)
	return m.resp, m.err
}

// makeRequest builds a minimal ExportTraceServiceRequest with a single resource
// span. If sessionID is non-empty, it is set as a resource attribute.
func makeRequest(sessionID string) *collectortrace.ExportTraceServiceRequest {
	var attrs []*commonpb.KeyValue
	if sessionID != "" {
		attrs = append(attrs, &commonpb.KeyValue{
			Key:   "session.id",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: sessionID}},
		})
	}
	return &collectortrace.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{Attributes: attrs},
				ScopeSpans: []*tracepb.ScopeSpans{
					{Spans: []*tracepb.Span{{Name: "test-span"}}},
				},
			},
		},
	}
}

func TestSessionIDExtraction(t *testing.T) {
	req := makeRequest("abc123")
	got := extractSessionID(req)
	if got != "abc123" {
		t.Errorf("extractSessionID: got %q, want %q", got, "abc123")
	}
}

func TestSessionIDExtractionMissing(t *testing.T) {
	req := makeRequest("")
	got := extractSessionID(req)
	if got != "" {
		t.Errorf("extractSessionID missing: got %q, want empty string", got)
	}
}

func TestSessionIDExtractionMultipleResources(t *testing.T) {
	// First resource has no session.id; second has "xyz". The last seen value
	// should be returned.
	req := &collectortrace.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key:   "service.name",
							Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "svc-a"}},
						},
					},
				},
			},
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key:   "session.id",
							Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "xyz"}},
						},
					},
				},
			},
		},
	}
	got := extractSessionID(req)
	if got != "xyz" {
		t.Errorf("extractSessionID multiple resources: got %q, want %q", got, "xyz")
	}
}

func TestProxyStoreSessionID(t *testing.T) {
	fwd := &mockForwarder{resp: &collectortrace.ExportTraceServiceResponse{}}
	p := New(fwd)

	if got := p.SessionID(); got != "" {
		t.Errorf("SessionID before any spans: got %q, want empty", got)
	}

	req := makeRequest("ses-99")
	_, err := p.Export(context.Background(), req)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if got := p.SessionID(); got != "ses-99" {
		t.Errorf("SessionID after export: got %q, want %q", got, "ses-99")
	}
}

func TestProxyForwardsSpans(t *testing.T) {
	wantResp := &collectortrace.ExportTraceServiceResponse{}
	fwd := &mockForwarder{resp: wantResp}
	p := New(fwd)

	req := makeRequest("fwd-test")
	got, err := p.Export(context.Background(), req)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if got != wantResp {
		t.Errorf("Export response: got %v, want %v", got, wantResp)
	}
	if len(fwd.calls) != 1 {
		t.Fatalf("forwarder called %d times, want 1", len(fwd.calls))
	}
	if fwd.calls[0] != req {
		t.Errorf("forwarder received different request than expected")
	}
}

func TestSessionIDNonStringAttributeIgnored(t *testing.T) {
	// Ensure a non-string AnyValue for session.id does not panic and returns "".
	req := &collectortrace.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "session.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_IntValue{IntValue: 42},
							},
						},
					},
				},
			},
		},
	}
	got := extractSessionID(req)
	if got != "" {
		t.Errorf("non-string session.id: got %q, want empty", got)
	}
}
