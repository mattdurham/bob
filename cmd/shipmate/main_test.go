package main

import (
	"strings"
	"testing"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantVals []string
	}{
		{"empty string", "", nil, nil},
		{"trailing commas", "Authorization=Bearer abc,,", []string{"Authorization"}, []string{"Bearer abc"}},
		{"missing equals", "BadEntry", nil, nil},
		{"valid pair", "Authorization=Bearer abc", []string{"Authorization"}, []string{"Bearer abc"}},
		{"multiple pairs", "X-Key=val1,Y-Key=val2", []string{"X-Key", "Y-Key"}, []string{"val1", "val2"}},
		{"value with equals sign", "Authorization=Bearer a=b", []string{"Authorization"}, []string{"Bearer a=b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHeaders(tt.input)
			for i, key := range tt.wantKeys {
				val, ok := got[key]
				if !ok {
					t.Errorf("key %q not found in headers", key)
					continue
				}
				if val != tt.wantVals[i] {
					t.Errorf("key %q: got %q, want %q", key, val, tt.wantVals[i])
				}
			}
			if len(tt.wantKeys) == 0 && len(got) != 0 {
				t.Errorf("expected empty headers, got %v", got)
			}
		})
	}
}

func TestParseHeadersNonNil(t *testing.T) {
	if parseHeaders("") == nil {
		t.Error("expected non-nil map for empty input")
	}
}

func TestSockPath(t *testing.T) {
	sessionID := "abc123"
	got, err := sockPath(sessionID)
	if err != nil {
		t.Fatalf("sockPath(%q) unexpected error: %v", sessionID, err)
	}
	if !strings.Contains(got, sessionID+".sock") {
		t.Errorf("sockPath(%q) = %q, want it to contain %q", sessionID, got, sessionID+".sock")
	}
}
