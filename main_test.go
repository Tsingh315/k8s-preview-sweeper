package main

import (
	"testing"
	"time"
)

func TestDecide(t *testing.T) {
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		annotations map[string]string
		want string
	}{
		{"no annotation", nil, "skip"},
		{"bad value", map[string]string{"expires-at": "tomorrow"}, "skip"},
		{"future", map[string]string{"expires-at": "2026-12-31T00:00:00Z"}, "keep"},
		{"past", map[string]string{"expires-at": "2026-10-01T00:00:00Z"}, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			got, reason := decide(tt.annotations, now)
			if got != tt.want {
				t.Errorf("decide() = %q, reason = %q, want %q", got, reason, tt.want)
			}
		})
	}
}
