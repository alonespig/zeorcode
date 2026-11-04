package judgeartifact

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/alonespig/zeorcode/backend/internal/repository"
)

const (
	manifestName      = "manifest.json"
	infoName          = "info.json"
	caseDirectory     = "cases"
	maxArtifactBytes  = int64(1 << 30)
	maxArtifactInputs = 2000
)

type Descriptor struct {
	ProblemID int64  `json:"problemId"`
	Version   string `json:"version"`
	Digest    string `json:"digest"`
}

type Store struct {
	root    string
	cacheMu sync.RWMutex
	cache   map[int64]Descriptor
}

func NewStore(root string) *Store {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "ojdata/problems"
	}
	return &Store{root: filepath.Clean(root), cache: make(map[int64]Descriptor)}
}

func (s *Store) Describe(ctx context.Context, problemID int64) (Descriptor, error) {
	if problemID <= 0 {
		return Descriptor{}, fmt.Errorf("invalid problem id %d", problemID)
	}
	dir := filepath.Join(s.root, strconv.FormatInt(problemID, 10))
	info, err := repository.LoadOrGenInfo(dir)
	if err != nil {
		return Descriptor{}, fmt.Errorf("load problem %d artifact info: %w", problemID, err)
	}
	s.cacheMu.RLock()
	cached, ok := s.cache[problemID]
	s.cacheMu.RUnlock()
	if ok && cached.Version == info.Version {
		return cached, nil
	}

	info, infoBytes, digest, err := describeDirectory(ctx, dir)
	_ = infoBytes
	if err != nil {
		return Descriptor{}, fmt.Errorf("describe problem %d artifact: %w", problemID, err)
	}
	descriptor := Descriptor{ProblemID: problemID, Version: info.Version, Digest: digest}
	s.cacheMu.Lock()
	s.cache[problemID] = descriptor
	s.cacheMu.Unlock()
	return descriptor, nil
}

func (s *Store) WriteArchive(ctx context.Context, w io.Writer, want Descriptor) error {
	if want.ProblemID <= 0 || want.Version == "" || !validDigest(want.Digest) {
		return errors.New("invalid artifact descriptor")
	}
	dir := filepath.Join(s.root, strconv.FormatInt(want.ProblemID, 10))
	info, infoBytes, digest, err := describeDirectory(ctx, dir)
	if err != nil {
		return err
	}
	if info.Version != want.Version || digest != want.Digest {
		return fmt.Errorf("artifact changed: current version=%s digest=%s", info.Version, digest)
	}

	zw := zip.NewWriter(w)
	closeWithErr := func(cause error) error {
		if closeErr := zw.Close(); cause == nil {
			return closeErr
		}
		return cause
	}
	manifestBytes, err := json.Marshal(want)
	if err != nil {
		return closeWithErr(err)
	}
	if err := writeZipBytes(zw, manifestName, manifestBytes); err != nil {
		return closeWithErr(err)
	}
	if err := writeZipBytes(zw, infoName, infoBytes); err != nil {
		return closeWithErr(err)
	}
	for _, tc := range sortedCases(info.Cases) {
		if err := ctx.Err(); err != nil {
			return closeWithErr(err)
		}
		if !validInputName(tc.Input) {
			return closeWithErr(fmt.Errorf("invalid testcase input name %q", tc.Input))
		}
		if err := writeZipFile(zw, filepath.Join(dir, tc.Input), caseDirectory+"/"+tc.Input); err != nil {
			return closeWithErr(err)
		}
	}
	return zw.Close()
}

func VerifyDirectory(ctx context.Context, dir string, want Descriptor) (*repository.TestCaseInfo, error) {
	info, _, digest, err := describeDirectory(ctx, dir)
	if err != nil {
		return nil, err
	}
	if info.Version != want.Version {
		return nil, fmt.Errorf("artifact version mismatch: got %s want %s", info.Version, want.Version)
	}
	if digest != want.Digest {
		return nil, fmt.Errorf("artifact digest mismatch: got %s want %s", digest, want.Digest)
	}
	return info, nil
}

func describeDirectory(ctx context.Context, dir string) (*repository.TestCaseInfo, []byte, string, error) {
	info, err := repository.LoadOrGenInfo(dir)
	if err != nil {
		return nil, nil, "", err
	}
	if info.Version == "" || info.Count <= 0 || len(info.Cases) != info.Count {
		return nil, nil, "", errors.New("invalid testcase info")
	}
	if len(info.Cases) > maxArtifactInputs {
		return nil, nil, "", fmt.Errorf("too many testcase inputs: %d", len(info.Cases))
	}
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return nil, nil, "", err
	}

	h := sha256.New()
	writeDigestPart(h, infoName, infoBytes)
	var total int64
	for _, tc := range sortedCases(info.Cases) {
		if err := ctx.Err(); err != nil {
			return nil, nil, "", err
		}
		if !validInputName(tc.Input) {
			return nil, nil, "", fmt.Errorf("invalid testcase input name %q", tc.Input)
		}
		data, err := os.ReadFile(filepath.Join(dir, tc.Input))
		if err != nil {
			return nil, nil, "", fmt.Errorf("read testcase input %q: %w", tc.Input, err)
		}
		total += int64(len(data))
		if total > maxArtifactBytes {
			return nil, nil, "", errors.New("artifact inputs exceed size limit")
		}
		writeDigestPart(h, caseDirectory+"/"+tc.Input, data)
	}
	return info, infoBytes, hex.EncodeToString(h.Sum(nil)), nil
}

func sortedCases(cases []repository.CaseInfo) []repository.CaseInfo {
	result := append([]repository.CaseInfo(nil), cases...)
	sort.Slice(result, func(i, j int) bool { return result[i].Input < result[j].Input })
	return result
}

func writeDigestPart(w io.Writer, name string, data []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(name)))
	_, _ = w.Write(size[:])
	_, _ = io.WriteString(w, name)
	binary.BigEndian.PutUint64(size[:], uint64(len(data)))
	_, _ = w.Write(size[:])
	_, _ = w.Write(data)
}

func validInputName(name string) bool {
	return name != "" && filepath.Base(name) == name && filepath.IsLocal(name) && strings.HasSuffix(name, ".in")
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func writeZipBytes(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func writeZipFile(zw *zip.Writer, source, name string) error {
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, io.LimitReader(f, maxArtifactBytes+1))
	return err
}
