package testdatastore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHasTestData(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{name: "missing directory", want: false},
		{name: "empty directory", files: []string{}, want: false},
		{name: "input only", files: []string{"1.in"}, want: false},
		{name: "output only", files: []string{"1.out"}, want: false},
		{name: "complete pair", files: []string{"sample.in", "sample.out"}, want: true},
		{name: "unmatched and complete pairs", files: []string{"1.in", "2.in", "2.out"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			problemID := int64(7)
			if tt.files != nil {
				dir := filepath.Join(root, "7")
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatalf("Mkdir() error = %v", err)
				}
				for _, name := range tt.files {
					if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0o644); err != nil {
						t.Fatalf("WriteFile(%q) error = %v", name, err)
					}
				}
			}

			got, err := NewAt(root).HasTestData(context.Background(), problemID)
			if err != nil {
				t.Fatalf("HasTestData() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("HasTestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasTestDataHonorsCanceledContext(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "7")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewAt(root).HasTestData(ctx, 7); err == nil {
		t.Fatal("HasTestData() error = nil, want canceled context error")
	}
}
