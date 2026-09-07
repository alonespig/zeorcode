package testdatastore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const defaultRoot = "ojdata/problems"

// Store inspects the filesystem-backed test data owned by local problems.
type Store struct {
	root string
}

// New creates a test-data store from judge.data_dir.
func New() *Store {
	return NewAt(viper.GetString("judge.data_dir"))
}

// NewAt creates a test-data store rooted at root. It is primarily useful for
// explicit process wiring and isolated tests.
func NewAt(root string) *Store {
	root = strings.TrimSpace(root)
	if root == "" {
		root = defaultRoot
	}
	return &Store{root: filepath.Clean(root)}
}

// HasTestData reports whether a problem has at least one readable .in/.out
// pair. A missing problem directory is a normal "not ready" state.
func (s *Store) HasTestData(ctx context.Context, problemID int64) (bool, error) {
	if problemID <= 0 {
		return false, fmt.Errorf("invalid problem id %d", problemID)
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	dir := filepath.Join(s.root, fmt.Sprintf("%d", problemID))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read test data directory %q: %w", dir, err)
	}

	files := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if !entry.IsDir() {
			files[entry.Name()] = struct{}{}
		}
	}

	hasPair := false
	for name := range files {
		var counterpart string
		switch {
		case strings.HasSuffix(name, ".in"):
			base := strings.TrimSuffix(name, ".in")
			if base == "" {
				return false, nil
			}
			counterpart = base + ".out"
		case strings.HasSuffix(name, ".out"):
			base := strings.TrimSuffix(name, ".out")
			if base == "" {
				return false, nil
			}
			counterpart = base + ".in"
		default:
			continue
		}

		if _, ok := files[counterpart]; !ok {
			return false, nil
		}
		if err := readableFile(filepath.Join(dir, name)); err != nil {
			return false, err
		}
		hasPair = true
	}

	return hasPair, nil
}

func readableFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open test data file %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close test data file %q: %w", path, err)
	}
	return nil
}
