package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTestcaseName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "1.in", want: true},
		{name: "1.out", want: true},
		{name: "sample-01.in", want: true},
		{name: "photo.jpg", want: false},
		{name: "archive.zip", want: false},
		{name: ".in", want: false},
		{name: "../1.in", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTestcaseName(tt.name); got != tt.want {
				t.Fatalf("isTestcaseName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestReadTestDataPreview(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("3 4\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	content, size, truncated, err := readTestDataPreview(dir, "1.in")
	if err != nil {
		t.Fatal(err)
	}
	if content != "3 4\n" || size != 4 || truncated {
		t.Fatalf("unexpected preview: content=%q size=%d truncated=%v", content, size, truncated)
	}

	if _, _, _, err := readTestDataPreview(dir, "../1.in"); err == nil {
		t.Fatal("expected traversal filename to be rejected")
	}
}

func TestReadTestDataPreviewTruncatesLargeFile(t *testing.T) {
	dir := t.TempDir()
	data := strings.Repeat("x", int(maxTestDataPreviewSize)+1)
	if err := os.WriteFile(filepath.Join(dir, "large.out"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	content, size, truncated, err := readTestDataPreview(dir, "large.out")
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(content)) != maxTestDataPreviewSize || size != int64(len(data)) || !truncated {
		t.Fatalf("unexpected truncated preview: len=%d size=%d truncated=%v", len(content), size, truncated)
	}
}
