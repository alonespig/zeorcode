package service

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestReadProblemArchiveAcceptsValidArchive(t *testing.T) {
	content := []byte("1 2\n")
	manifest := testProblemArchiveManifest(content)
	zr := buildProblemArchive(t, manifest, map[string][]byte{
		"data/000001/1.in": content,
	})

	parsed, err := readProblemArchive(zr)
	if err != nil {
		t.Fatalf("readProblemArchive() error = %v", err)
	}
	if len(parsed.manifest.Problems) != 1 || parsed.manifest.Problems[0].DisplayID != "P1001" {
		t.Fatalf("unexpected manifest: %+v", parsed.manifest)
	}
}

func TestReadProblemArchiveAcceptsEmptyArchive(t *testing.T) {
	manifest := ProblemArchiveManifest{
		Format:     ProblemArchiveFormat,
		Version:    ProblemArchiveVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Problems:   []ProblemArchiveProblem{},
	}
	zr := buildProblemArchive(t, manifest, nil)

	parsed, err := readProblemArchive(zr)
	if err != nil {
		t.Fatalf("readProblemArchive() error = %v", err)
	}
	if len(parsed.manifest.Problems) != 0 {
		t.Fatalf("unexpected problems: %+v", parsed.manifest.Problems)
	}
}

func TestReadProblemArchiveRejectsPathTraversal(t *testing.T) {
	manifest := testProblemArchiveManifest(nil)
	manifest.Problems[0].TestData = nil
	zr := buildProblemArchive(t, manifest, map[string][]byte{
		"../escape.in": []byte("unsafe"),
	})

	if _, err := readProblemArchive(zr); err == nil || !strings.Contains(err.Error(), "非法路径") {
		t.Fatalf("readProblemArchive() error = %v, want invalid path", err)
	}
}

func TestExtractArchivedTestDataRejectsChecksumMismatch(t *testing.T) {
	content := []byte("1 2\n")
	manifest := testProblemArchiveManifest(content)
	manifest.Problems[0].TestData[0].SHA256 = strings.Repeat("0", sha256.Size*2)
	zr := buildProblemArchive(t, manifest, map[string][]byte{
		"data/000001/1.in": content,
	})
	parsed, err := readProblemArchive(zr)
	if err != nil {
		t.Fatalf("readProblemArchive() error = %v", err)
	}

	if err := extractArchivedTestData(parsed, 0, t.TempDir()); err == nil || !strings.Contains(err.Error(), "校验失败") {
		t.Fatalf("extractArchivedTestData() error = %v, want checksum failure", err)
	}
}

func testProblemArchiveManifest(content []byte) ProblemArchiveManifest {
	hash := sha256.Sum256(content)
	return ProblemArchiveManifest{
		Format:     ProblemArchiveFormat,
		Version:    ProblemArchiveVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Problems: []ProblemArchiveProblem{{
			DisplayID:   "P1001",
			Name:        "A + B",
			Difficulty:  1,
			TimeLimit:   1000,
			MemoryLimit: 256,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
			Tags:        []string{"入门"},
			Samples:     []ProblemArchiveSample{},
			TestData: []ProblemArchiveFile{{
				Name: "1.in", Size: int64(len(content)), SHA256: hex.EncodeToString(hash[:]),
			}},
		}},
	}
}

func buildProblemArchive(t *testing.T, manifest ProblemArchiveManifest, files map[string][]byte) *zip.Reader {
	t.Helper()
	var buffer bytes.Buffer
	zw := zip.NewWriter(&buffer)
	for name, content := range files {
		writer, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := writer.Write(content); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	manifestWriter, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	if err := json.NewEncoder(manifestWriter).Encode(manifest); err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	return zr
}
