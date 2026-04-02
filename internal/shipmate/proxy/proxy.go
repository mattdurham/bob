package proxy

import (
	"context"
	"net"
	"sync"
	"time"

	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/grpc"
)

// Forwarder is implemented by anything that can forward a raw export request.
// Using an interface keeps proxy testable without a real upstream.
type Forwarder interface {
	Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error)
}

// Proxy is a gRPC OTLP receiver. It extracts session.id from incoming spans
// and forwards the request to the upstream Forwarder unchanged.
type Proxy struct {
	collectortrace.UnimplementedTraceServiceServer

	mu        sync.RWMutex
	sessionID string
	forwarder Forwarder
}

// New creates a Proxy with the given forwarder.
func New(fwd Forwarder) *Proxy {
	return &Proxy{forwarder: fwd}
}

// SessionID returns the most recently observed session.id from incoming spans.
// Returns empty string if no spans have been received yet.
func (p *Proxy) SessionID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sessionID
}

// Export implements TraceServiceServer. It extracts session.id, updates state,
// then forwards the request to the upstream.
func (p *Proxy) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	if id := extractSessionID(req); id != "" {
		p.mu.Lock()
		p.sessionID = id
		p.mu.Unlock()
	}
	return p.forwarder.Export(ctx, req)
}

// extractSessionID scans all resource spans for the "session.id" resource attribute.
// Returns the value from the last resource span that has it, or empty string.
func extractSessionID(req *collectortrace.ExportTraceServiceRequest) string {
	var found string
	for _, rs := range req.GetResourceSpans() {
		attrs := rs.GetResource().GetAttributes()
		for _, kv := range attrs {
			if kv.GetKey() == "session.id" {
				found = stringValue(kv.GetValue())
				break // only one session.id expected per resource
			}
		}
	}
	return found
}

func stringValue(v *commonpb.AnyValue) string {
	if v == nil {
		return ""
	}
	if sv, ok := v.GetValue().(*commonpb.AnyValue_StringValue); ok {
		return sv.StringValue
	}
	return ""
}

// Listen starts the gRPC server on the given address and blocks until ctx is done.
func (p *Proxy) Listen(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(srv, p)

	go func() {
		<-ctx.Done()
		stopped := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			srv.Stop()
		}
	}()

	return srv.Serve(lis)
}
