// shipmate is a stdio MCP server that acts as an enriching OTLP proxy.
// It receives OTEL spans from Claude Code via gRPC, forwards them to a
// downstream OTLP HTTP endpoint, and lets agents emit synthetic spans via
// the shipmate_record MCP tool.
//
// Configuration (env vars):
//
//	SHIPMATE_OTLP_LISTEN_ADDR   gRPC listen address (default :4317)
//	SHIPMATE_UPSTREAM_ENDPOINT  Required. Upstream OTLP HTTP endpoint, e.g. http://localhost:4318
//	SHIPMATE_UPSTREAM_HEADERS   Optional. Comma-separated Key=Value auth headers.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/proxy"
	"github.com/mattdurham/bob/internal/shipmate/recorder"
	"github.com/mattdurham/bob/internal/shipmate/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
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

	upstreamHeaders := parseHeaders(headersRaw)

	// Build proxy forwarder (gRPC receive → HTTP forward).
	fwd := newHTTPForwarder(upstreamEndpoint, upstreamHeaders)
	prx := proxy.New(fwd)

	// Build OTLP HTTP exporter for synthetic spans.
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(upstreamEndpoint),
		otlptracehttp.WithHeaders(upstreamHeaders),
	}
	if !strings.HasPrefix(upstreamEndpoint, "https://") {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		log.Fatalf("shipmate: create exporter: %v", err)
	}

	rec, err := recorder.New(exp)
	if err != nil {
		log.Fatalf("shipmate: create recorder: %v", err)
	}
	defer func() {
		if err := rec.Shutdown(context.Background()); err != nil {
			log.Printf("shipmate: recorder shutdown: %v", err)
		}
	}()

	// Start gRPC proxy listener; cancel context on failure.
	go func() {
		if err := prx.Listen(ctx, listenAddr); err != nil {
			log.Printf("shipmate: proxy stopped: %v", err)
			cancel()
		}
	}()
	log.Printf("shipmate: OTLP receiver listening on %s", listenAddr)

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

// httpForwarder implements proxy.Forwarder by POSTing the proto-encoded
// request to the upstream's /v1/traces endpoint over HTTP.
type httpForwarder struct {
	endpoint string
	headers  map[string]string
	client   *http.Client
}

func newHTTPForwarder(endpoint string, headers map[string]string) *httpForwarder {
	return &httpForwarder{
		endpoint: strings.TrimRight(endpoint, "/"),
		headers:  headers,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *httpForwarder) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	body, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint+"/v1/traces", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	for k, v := range f.headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

// parseHeaders parses "Key=Value,Key2=Value2" into a string map.
// Malformed pairs are silently skipped.
func parseHeaders(raw string) map[string]string {
	headers := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return headers
}
