package artifact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	info := TestCaseInfo{
		Version: "version-1",
		Mode:    "default",
		Count:   1,
		Cases: []CaseInfo{{
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
	input := []byte("2 3\n")
	if err := os.WriteFile(filepath.Join(dir, "info.json"), infoBytes, 0o640); err != nil {
		t.Fatalf("write info: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1.in"), input, 0o640); err != nil {
		t.Fatalf("write input: %v", err)
	}
	h := sha256.New()
	writeDigestPart(h, "info.json", infoBytes)
	writeDigestPart(h, "cases/1.in", input)
	want := Descriptor{ProblemID: 1, Version: info.Version, Digest: hex.EncodeToString(h.Sum(nil))}

	got, err := VerifyDirectory(context.Background(), dir, want)
	if err != nil {
		t.Fatalf("verify directory: %v", err)
	}
	if got.Count != 1 || got.Cases[0].Input != "1.in" {
		t.Fatalf("unexpected info: %+v", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("corrupted\n"), 0o640); err != nil {
		t.Fatalf("corrupt input: %v", err)
	}
	if _, err := VerifyDirectory(context.Background(), dir, want); err == nil {
		t.Fatal("expected corrupted input to fail digest verification")
	}
}
