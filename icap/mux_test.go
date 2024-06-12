package icap

import (
	"testing"
)

func TestNewServeMux(t *testing.T) {
	mux := NewServeMux()
	if mux == nil {
		t.Fatal("NewServeMux() returned nil")
	}
	if len(mux.m) != 0 {
		t.Errorf("Expected empty handler map, got %d handlers", len(mux.m))
	}
}

func TestPathMatch(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"/", "/", true},
		{"/foo", "/foo", true},
		{"/foo/", "/foo/bar", true},
		{"/foo", "/foo/bar", false},
		{"/foo/", "/foo", false},
		{"/foo/bar", "/foo", false},
		{"", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			if got := pathMatch(tt.pattern, tt.path); got != tt.want {
				t.Errorf("pathMatch(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
