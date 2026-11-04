package judgenode

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/alonespig/zeorcode/judge-node/internal/artifact"
)

func TestArtifactCacheDownloadsAndRepairsArtifact(t *testing.T) {
	t.Parallel()

	descriptor, archive := testArchive(t)
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
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	cache, err := NewArtifactCache(t.TempDir(), server.URL, token)
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	dir, info, err := cache.Ensure(context.Background(), descriptor)
	if err != nil {
		t.Fatalf("download artifact: %v", err)
	}
	if info.Count != 1 {
		t.Fatalf("testcase count = %d, want 1", info.Count)
	}
	if _, err := os.Stat(filepath.Join(dir, "1.in")); err != nil {
		t.Fatalf("cached input missing: %v", err)
	}
	if _, _, err := cache.Ensure(context.Background(), descriptor); err != nil {
		t.Fatalf("reuse cached artifact: %v", err)
	}
	if got := downloads.Load(); got != 1 {
		t.Fatalf("cache hit downloaded again: %d", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("corrupted\n"), 0o640); err != nil {
		t.Fatalf("corrupt cached input: %v", err)
	}
	if _, _, err := cache.Ensure(context.Background(), descriptor); err != nil {
		t.Fatalf("repair cached artifact: %v", err)
	}
	if got := downloads.Load(); got != 2 {
		t.Fatalf("corrupt cache downloads = %d, want 2", got)
	}
}

func testArchive(t *testing.T) (artifact.Descriptor, []byte) {
	t.Helper()
	info := artifact.TestCaseInfo{
		Version: "version-1",
		Mode:    "default",
		Count:   1,
		Cases: []artifact.CaseInfo{{
			ID:             1,
			Input:          "1.in",
			Output:         "1.out",
			StrippedMd5:    "expected-hash",
			AllStrippedMd5: "expected-space-hash",
		}},
	}
	infoBytes, err := json.Marshal(&info)
	if err != nil {
		t.Fatalf("marshal info: %v", err)
	}
	input := []byte("4 5\n")
	h := sha256.New()
	testDigestPart(h, "info.json", infoBytes)
	testDigestPart(h, "cases/1.in", input)
	descriptor := artifact.Descriptor{
		ProblemID: 202,
		Version:   info.Version,
		Digest:    hex.EncodeToString(h.Sum(nil)),
	}
	manifest, err := json.Marshal(&descriptor)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	var result bytes.Buffer
	zw := zip.NewWriter(&result)
	for name, content := range map[string][]byte{
		"manifest.json": manifest,
		"info.json":     infoBytes,
		"cases/1.in":    input,
	} {
		entry, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return descriptor, result.Bytes()
}

func testDigestPart(w io.Writer, name string, data []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(name)))
	_, _ = w.Write(size[:])
	_, _ = io.WriteString(w, name)
	binary.BigEndian.PutUint64(size[:], uint64(len(data)))
	_, _ = w.Write(size[:])
	_, _ = w.Write(data)
}
