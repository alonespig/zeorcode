package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zoj/pkg/judge"

	"github.com/spf13/viper"
)

const defaultRoot = "ojdata/problems"

// ProblemTestDataStore inspects the filesystem-backed test data owned by local problems.
type ProblemTestDataStore struct {
	root string
}

// NewProblemTestDataStore creates a test-data store from judge.data_dir.
func NewProblemTestDataStore() *ProblemTestDataStore {
	return NewProblemTestDataStoreAt(viper.GetString("judge.data_dir"))
}

// NewProblemTestDataStoreAt creates a test-data store rooted at root. It is primarily useful for
// explicit process wiring and isolated tests.
func NewProblemTestDataStoreAt(root string) *ProblemTestDataStore {
	root = strings.TrimSpace(root)
	if root == "" {
		root = defaultRoot
	}
	return &ProblemTestDataStore{root: filepath.Clean(root)}
}

// HasTestData reports whether a problem has at least one readable .in/.out
// pair. A missing problem directory is a normal "not ready" state.
func (s *ProblemTestDataStore) HasTestData(ctx context.Context, problemID int64) (bool, error) {
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

// info.json：每题测试数据清单（测试点 → 输入/输出文件 + 输出哈希），
// 放在该题测试数据目录下。判题时按哈希比对，支持 AC / PE / WA 三档。
const infoFileName = "info.json"

// CaseInfo 单个测试点的清单信息
type CaseInfo struct {
	ID             int    `json:"id"`
	Input          string `json:"input"`          // 输入文件名，如 1.in
	Output         string `json:"output"`         // 输出文件名，如 1.out
	OutputSize     int    `json:"outputSize"`     // 输出字节数
	OutputMd5      string `json:"outputMd5"`      // 原样输出 md5（精确比）
	StrippedMd5    string `json:"strippedMd5"`    // 去每行行末 + 文末空白(rtrim) 后 md5 —— 判 AC
	AllStrippedMd5 string `json:"allStrippedMd5"` // 去掉所有空白后 md5 —— 判 PE
}

// TestCaseInfo 一题的测试数据清单
type TestCaseInfo struct {
	Version string     `json:"version"` // 版本戳，数据变更后重新生成
	Mode    string     `json:"mode"`    // default / spj / interactive（暂固定 default）
	Count   int        `json:"count"`
	Cases   []CaseInfo `json:"cases"`
}

// generateInfo scans basename-matched .in/.out pairs, calculates hashes, and
// writes info.json. Case IDs are stable for a given filename ordering and do
// not require the basename itself to be numeric.
func generateInfo(testDir string) (*TestCaseInfo, error) {
	entries, err := os.ReadDir(testDir)
	if err != nil {
		return nil, err
	}
	files := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files[entry.Name()] = struct{}{}
		}
	}
	for name := range files {
		if !strings.HasSuffix(name, ".out") {
			continue
		}
		base := strings.TrimSuffix(name, ".out")
		if base == "" {
			return nil, fmt.Errorf("invalid testcase output name %q", name)
		}
		if _, ok := files[base+".in"]; !ok {
			return nil, fmt.Errorf("testcase output %q has no matching input", name)
		}
	}
	cases := make([]CaseInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".in") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".in")
		if base == "" {
			return nil, fmt.Errorf("invalid testcase input name %q", e.Name())
		}
		outName := base + ".out"
		outBytes, err := os.ReadFile(filepath.Join(testDir, outName))
		if err != nil {
			return nil, fmt.Errorf("read testcase output %q: %w", outName, err)
		}
		out := string(outBytes)
		cases = append(cases, CaseInfo{
			Input:          e.Name(),
			Output:         outName,
			OutputSize:     len(outBytes),
			OutputMd5:      judge.MD5(out),
			StrippedMd5:    judge.MD5(judge.RtrimOutput(out)),
			AllStrippedMd5: judge.MD5(judge.StripAllSpace(out)),
		})
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Input < cases[j].Input })
	for i := range cases {
		cases[i].ID = i + 1
	}

	info := &TestCaseInfo{
		Version: strconv.FormatInt(time.Now().Unix(), 10),
		Mode:    "default",
		Count:   len(cases),
		Cases:   cases,
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(testDir, infoFileName), data, 0o644); err != nil {
		return nil, err
	}
	return info, nil
}

// RebuildInfo 立即扫描目录、算哈希并写出 info.json，供上传/改动测试数据后调用。
// 生成失败时删除旧清单，避免留下过期哈希。
func RebuildInfo(testDir string) error {
	if _, err := generateInfo(testDir); err != nil {
		_ = os.Remove(filepath.Join(testDir, infoFileName))
		return err
	}
	return nil
}

// LoadOrGenInfo 读 info.json；不存在或损坏则扫描目录重新生成（懒生成）。
func LoadOrGenInfo(testDir string) (*TestCaseInfo, error) {
	data, err := os.ReadFile(filepath.Join(testDir, infoFileName))
	if err != nil {
		return generateInfo(testDir)
	}
	var info TestCaseInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return generateInfo(testDir)
	}
	return &info, nil
}
