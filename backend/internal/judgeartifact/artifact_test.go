package judgeartifact

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alonespig/zeorcode/backend/internal/repository"
)

func TestStoreDescribeAndArchive(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	problemDir := filepath.Join(root, "101")
	writeTestCase(t, problemDir, "1", "2 3\n", "5\n")
	writeTestCase(t, problemDir, "2", "7 8\n", "15\n")
	if err := repository.RebuildInfo(problemDir); err != nil {
		t.Fatalf("rebuild info: %v", err)
	}

	store := NewStore(root)
	descriptor, err := store.Describe(context.Background(), 101)
	if err != nil {
		t.Fatalf("describe artifact: %v", err)
	}
	if descriptor.Version == "" || len(descriptor.Digest) != 64 {
		t.Fatalf("unexpected descriptor: %+v", descriptor)
	}

	var archive bytes.Buffer
	if err := store.WriteArchive(context.Background(), &archive, descriptor); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	names := make(map[string]bool, len(zr.File))
	for _, file := range zr.File {
		names[file.Name] = true
	}
	for _, want := range []string{"manifest.json", "info.json", "cases/1.in", "cases/2.in"} {
		if !names[want] {
			t.Errorf("archive is missing %s", want)
		}
	}
	if names["cases/1.out"] || names["1.out"] {
		t.Fatal("artifact must not contain expected outputs")
	}

	if err := os.WriteFile(filepath.Join(problemDir, "1.in"), []byte("changed\n"), 0o640); err != nil {
		t.Fatalf("change input: %v", err)
	}
	if err := repository.RebuildInfo(problemDir); err != nil {
		t.Fatalf("rebuild changed info: %v", err)
	}
	changed, err := store.Describe(context.Background(), 101)
	if err != nil {
		t.Fatalf("describe changed artifact: %v", err)
	}
	if changed.Digest == descriptor.Digest {
		t.Fatal("input change did not change artifact digest")
	}
	if err := store.WriteArchive(context.Background(), &bytes.Buffer{}, descriptor); err == nil {
		t.Fatal("expected stale descriptor to be rejected")
	}
}

func writeTestCase(t *testing.T, dir, name, input, output string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create testcase directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".in"), []byte(input), 0o640); err != nil {
		t.Fatalf("write testcase input: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".out"), []byte(output), 0o640); err != nil {
		t.Fatalf("write testcase output: %v", err)
	}
}
