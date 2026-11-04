package judgenode

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/alonespig/zeorcode/backend/internal/judgeartifact"
	"github.com/alonespig/zeorcode/backend/internal/repository"
)

func TestArtifactCacheDownloadsAndRepairsArtifact(t *testing.T) {
	t.Parallel()

	sourceRoot := t.TempDir()
	problemID := int64(202)
	sourceDir := filepath.Join(sourceRoot, strconv.FormatInt(problemID, 10))
	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("create source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "1.in"), []byte("4 5\n"), 0o640); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "1.out"), []byte("9\n"), 0o640); err != nil {
		t.Fatalf("write output: %v", err)
	}
	if err := repository.RebuildInfo(sourceDir); err != nil {
		t.Fatalf("rebuild info: %v", err)
	}
	store := judgeartifact.NewStore(sourceRoot)
	descriptor, err := store.Describe(context.Background(), problemID)
	if err != nil {
		t.Fatalf("describe source artifact: %v", err)
	}

	const token = "artifact-test-token"
	var downloads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downloads.Add(1)
		if r.URL.Path != "/api/internal/judge/problems/202/artifact" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("version") != descriptor.Version || r.URL.Query().Get("digest") != descriptor.Digest {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		if err := store.WriteArchive(r.Context(), w, descriptor); err != nil {
			t.Errorf("write artifact response: %v", err)
		}
	}))
	defer server.Close()

	cache, err := NewArtifactCache(t.TempDir(), server.URL, token)
	if err != nil {
		t.Fatalf("create artifact cache: %v", err)
	}
	dir, info, err := cache.Ensure(context.Background(), descriptor)
	if err != nil {
		t.Fatalf("download artifact: %v", err)
	}
	if info.Count != 1 {
		t.Fatalf("unexpected testcase count: %d", info.Count)
	}
	if _, err := os.Stat(filepath.Join(dir, "1.in")); err != nil {
		t.Fatalf("cached input missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "1.out")); !os.IsNotExist(err) {
		t.Fatal("cached artifact unexpectedly contains expected output")
	}
	if _, _, err := cache.Ensure(context.Background(), descriptor); err != nil {
		t.Fatalf("reuse cached artifact: %v", err)
	}
	if got := downloads.Load(); got != 1 {
		t.Fatalf("cache hit downloaded artifact again: downloads=%d", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("corrupted\n"), 0o640); err != nil {
		t.Fatalf("corrupt cache: %v", err)
	}
	if _, _, err := cache.Ensure(context.Background(), descriptor); err != nil {
		t.Fatalf("repair cached artifact: %v", err)
	}
	if got := downloads.Load(); got != 2 {
		t.Fatalf("corrupt cache was not downloaded again: downloads=%d", got)
	}
}
