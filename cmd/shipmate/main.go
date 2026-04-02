// shipmate is a stdio MCP server that acts as an enriching OTLP proxy.
// It receives OTEL spans from Claude Code, forwards them to a downstream
// OTLP endpoint, and lets agents emit synthetic spans via the shipmate_record
// MCP tool.
//
// Configuration (env vars):
//
//	SHIPMATE_OTLP_LISTEN_ADDR   gRPC listen address (default :4317)
//	SHIPMATE_UPSTREAM_ENDPOINT  Required. Upstream OTLP endpoint, e.g. localhost:14317
//	SHIPMATE_UPSTREAM_HEADERS   Optional. Comma-separated Key=Value auth headers.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mattdurham/bob/internal/shipmate/proxy"
	"github.com/mattdurham/bob/internal/shipmate/recorder"
	"github.com/mattdurham/bob/internal/shipmate/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	upstreamEndpoint := os.Getenv("SHIPMATE_UPSTREAM_ENDPOINT")
	if upstreamEndpoint == "" {
		log.Fatal("shipmate: SHIPMATE_UPSTREAM_ENDPOINT is required")
	}
	listenAddr := os.Getenv("SHIPMATE_OTLP_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":4317"
	}
	headersRaw := os.Getenv("SHIPMATE_UPSTREAM_HEADERS")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Parse headers once at startup; they are injected per-call in grpcForwarder.Export.
	upstreamHeaders := parseHeaders(headersRaw)

	// Build dial options for upstream gRPC connection.
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Dial the upstream once; share the connection between the raw forwarder
	// and the OTEL SDK exporter to avoid opening two connections.
	conn, err := grpc.NewClient(upstreamEndpoint, dialOpts...)
	if err != nil {
		log.Fatalf("shipmate: dial upstream: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("shipmate: close upstream conn: %v", err)
		}
	}()

	// Build the forwarder and proxy (gRPC OTLP receiver).
	fwd := newGRPCForwarder(conn, upstreamHeaders)
	prx := proxy.New(fwd)

	// Build the OTEL SDK exporter for synthetic spans, reusing the same conn.
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("shipmate: create exporter: %v", err)
	}
	// Build the recorder (synthetic spans).
	// The TracerProvider owns the exporter's lifetime; shutting down the provider is sufficient.
	rec, err := recorder.New(exp)
	if err != nil {
		log.Fatalf("shipmate: create recorder: %v", err)
	}
	defer func() {
		if err := rec.Shutdown(context.Background()); err != nil {
			log.Printf("shipmate: recorder shutdown: %v", err)
		}
	}()

	// Start gRPC proxy listener in a goroutine; it blocks until ctx is done.
	// If Listen fails (e.g. port in use), cancel the context to shut down the whole process.
	go func() {
		if err := prx.Listen(ctx, listenAddr); err != nil {
			log.Printf("shipmate: proxy stopped: %v", err)
			cancel()
		}
	}()
	log.Printf("shipmate: OTLP receiver listening on %s", listenAddr)

	// Build and run the MCP server on stdio (blocks until stdin closes).
	toolServer := server.New(rec, prx)
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "shipmate",
		Version: "v0.1.0",
	}, nil)
	toolServer.Register(srv)

	log.Printf("shipmate: MCP server starting (stdio)")
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("shipmate: %v", err)
	}
}

// grpcForwarder implements proxy.Forwarder using the generated gRPC client.
// It performs transparent forwarding: the raw proto request is sent to the
// upstream as-is, without any conversion through the OTEL SDK.
// Headers are injected explicitly per-call so injection is visible and
// independent of the shared connection's interceptor chain.
type grpcForwarder struct {
	client  collectortrace.TraceServiceClient
	headers metadata.MD
}

func newGRPCForwarder(conn *grpc.ClientConn, headers metadata.MD) *grpcForwarder {
	return &grpcForwarder{client: collectortrace.NewTraceServiceClient(conn), headers: headers}
}

func (f *grpcForwarder) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	outCtx := ctx
	if len(f.headers) > 0 {
		outCtx = metadata.NewOutgoingContext(ctx, f.headers)
	}
	return f.client.Export(outCtx, req)
}

// parseHeaders parses "Key=Value,Key2=Value2" into gRPC metadata.
// Keys are lowercased per gRPC convention. Malformed pairs are silently skipped.
func parseHeaders(raw string) metadata.MD {
	md := metadata.MD{}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		md[key] = []string{val}
	}
	return md
}
