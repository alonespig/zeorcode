package judgeworker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zoj/pkg/judge"
	"zoj/pkg/util"
)

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

// rtrimOutput 去掉每行行末空白 + 文末空白（容忍行末空格、末尾多余换行）
func rtrimOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t\f\v")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// stripAllSpace 去掉所有空白字符（用于 PE 判定：内容一致仅空白不同）
func stripAllSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// judgeOutput 用清单哈希判定用户输出：AC / PE / WA
func judgeOutput(stdout, strippedMd5, allStrippedMd5 string) int {
	if util.MD5(rtrimOutput(stdout)) == strippedMd5 {
		return judge.Accepted
	}
	if util.MD5(stripAllSpace(stdout)) == allStrippedMd5 {
		return judge.PresentationError
	}
	return judge.WrongAnswer
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
			OutputMd5:      util.MD5(out),
			StrippedMd5:    util.MD5(rtrimOutput(out)),
			AllStrippedMd5: util.MD5(stripAllSpace(out)),
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

// GenerateInfo 立即扫描目录、算哈希并写出 info.json。
// 供上传/改动测试数据后调用：清单在上传完即生成，而非等首次判题才懒生成。
func GenerateInfo(testDir string) error {
	_, err := generateInfo(testDir)
	return err
}

// loadOrGenInfo 读 info.json；不存在或损坏则扫描目录重新生成（懒生成）
func loadOrGenInfo(testDir string) (*TestCaseInfo, error) {
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
