package judgeworker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"zoj/internal/model"
	"zoj/pkg/judge"
)

func TestRunAllCasesRejectsEmptyCases(t *testing.T) {
	worker := &Worker{}
	_, err := worker.runAllCases(
		context.Background(),
		&model.Submission{},
		nil,
		nil,
		nil,
		judge.Limits{},
		nil,
	)
	if err == nil {
		t.Fatal("runAllCases() error = nil, want empty test data error")
	}
}

func TestLoadCasesFromDirRejectsMissingOrEmptyData(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		if _, err := loadCasesFromDir(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Fatal("loadCasesFromDir() error = nil, want missing data error")
		}
	})

	t.Run("no complete pairs", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("input"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if _, err := loadCasesFromDir(dir); err == nil {
			t.Fatal("loadCasesFromDir() error = nil, want empty data error")
		}
	})
}

func TestLoadCasesFromDirLoadsCompletePair(t *testing.T) {
	dir := t.TempDir()
	writeTestcaseFile(t, dir, "sample.in", "1 2\n")
	writeTestcaseFile(t, dir, "sample.out", "3\n")

	cases, err := loadCasesFromDir(dir)
	if err != nil {
		t.Fatalf("loadCasesFromDir() error = %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("len(cases) = %d, want 1", len(cases))
	}
	if cases[0].Input != "1 2\n" {
		t.Fatalf("case input = %q, want %q", cases[0].Input, "1 2\n")
	}
}

func TestLoadCasesFromDirRejectsMissingDeclaredInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "1.in")
	writeTestcaseFile(t, dir, "1.in", "input")
	writeTestcaseFile(t, dir, "1.out", "output")
	if err := GenerateInfo(dir); err != nil {
		t.Fatalf("GenerateInfo() error = %v", err)
	}
	if err := os.Remove(inputPath); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := loadCasesFromDir(dir); err == nil {
		t.Fatal("loadCasesFromDir() error = nil, want missing declared input error")
	}
}

func writeTestcaseFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", name, err)
	}
}
