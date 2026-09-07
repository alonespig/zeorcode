package service

import (
	"strings"
	"testing"
)

func TestNormalizeTagName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims spaces", input: "  动态规划  ", want: "动态规划"},
		{name: "empty", input: "  ", wantErr: true},
		{name: "too long", input: strings.Repeat("算", 65), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTagName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeTagName(%q) error = nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeTagName(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeTagName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
