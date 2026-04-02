package main

import (
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantVals []string
	}{
		{"empty string", "", nil, nil},
		{"trailing commas", "Authorization=Bearer abc,,", []string{"authorization"}, []string{"Bearer abc"}},
		{"missing equals", "BadEntry", nil, nil},
		{"valid pair", "Authorization=Bearer abc", []string{"authorization"}, []string{"Bearer abc"}},
		{"multiple pairs", "X-Key=val1,Y-Key=val2", []string{"x-key", "y-key"}, []string{"val1", "val2"}},
		{"value with equals sign", "Authorization=Bearer a=b", []string{"authorization"}, []string{"Bearer a=b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := parseHeaders(tt.input)
			for i, key := range tt.wantKeys {
				vals, ok := md[key]
				if !ok {
					t.Errorf("key %q not found in metadata", key)
					continue
				}
				if vals[0] != tt.wantVals[i] {
					t.Errorf("key %q: got %q, want %q", key, vals[0], tt.wantVals[i])
				}
			}
			// No extra keys for empty/bad inputs
			if len(tt.wantKeys) == 0 && len(md) != 0 {
				t.Errorf("expected empty metadata, got %v", md)
			}
		})
	}
}

// Ensure parseHeaders returns a non-nil metadata.MD even for empty input.
func TestParseHeadersReturnsNonNilForEmpty(t *testing.T) {
	md := parseHeaders("")
	if md == nil {
		t.Error("expected non-nil metadata.MD for empty input")
	}
	_ = metadata.MD(md) // compile-time type check
}
