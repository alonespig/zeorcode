package artifact

import (
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
	"strings"
)

const MaxBytes = int64(1 << 30)

type Descriptor struct {
	ProblemID int64  `json:"problemId"`
	Version   string `json:"version"`
	Digest    string `json:"digest"`
}

type CaseInfo struct {
	ID             int    `json:"id"`
	Input          string `json:"input"`
	Output         string `json:"output"`
	OutputSize     int    `json:"outputSize"`
	OutputMd5      string `json:"outputMd5"`
	StrippedMd5    string `json:"strippedMd5"`
	AllStrippedMd5 string `json:"allStrippedMd5"`
}

type TestCaseInfo struct {
	Version string     `json:"version"`
	Mode    string     `json:"mode"`
	Count   int        `json:"count"`
	Cases   []CaseInfo `json:"cases"`
}

func VerifyDirectory(ctx context.Context, dir string, want Descriptor) (*TestCaseInfo, error) {
	infoBytes, err := os.ReadFile(filepath.Join(dir, "info.json"))
	if err != nil {
		return nil, fmt.Errorf("read testcase info: %w", err)
	}
	var info TestCaseInfo
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return nil, errors.New("invalid testcase info")
	}
	if info.Version == "" || info.Count <= 0 || len(info.Cases) != info.Count {
		return nil, errors.New("invalid testcase info")
	}
	if info.Version != want.Version {
		return nil, fmt.Errorf("artifact version mismatch: got %s want %s", info.Version, want.Version)
	}
	canonicalInfo, err := json.Marshal(&info)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	writeDigestPart(h, "info.json", canonicalInfo)
	var total int64
	for _, tc := range sortedCases(info.Cases) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !ValidInputName(tc.Input) {
			return nil, fmt.Errorf("invalid testcase input name %q", tc.Input)
		}
		input, err := os.ReadFile(filepath.Join(dir, tc.Input))
		if err != nil {
			return nil, fmt.Errorf("read testcase input %q: %w", tc.Input, err)
		}
		total += int64(len(input))
		if total > MaxBytes {
			return nil, errors.New("artifact inputs exceed size limit")
		}
		writeDigestPart(h, "cases/"+tc.Input, input)
	}
	if digest := hex.EncodeToString(h.Sum(nil)); digest != want.Digest {
		return nil, fmt.Errorf("artifact digest mismatch: got %s want %s", digest, want.Digest)
	}
	return &info, nil
}

func ValidInputName(name string) bool {
	return name != "" && filepath.Base(name) == name && filepath.IsLocal(name) && strings.HasSuffix(name, ".in")
}

func ValidDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sortedCases(cases []CaseInfo) []CaseInfo {
	result := append([]CaseInfo(nil), cases...)
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
