package hdu

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html/charset"
)

// 读 GB2312 样本并转 UTF-8(与 fetchHTML 的转码一致)
func loadUTF8(t *testing.T, path string) []byte {
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读样本失败: %v", err)
	}
	r, err := charset.NewReader(bytes.NewReader(raw), "text/html; charset=gb2312")
	if err != nil {
		t.Fatalf("转码失败: %v", err)
	}
	b, _ := io.ReadAll(r)
	return b
}

func TestParseProblem(t *testing.T) {
	utf8Bytes := loadUTF8(t, "testdata/p7836.html")
	p, err := parseProblem(utf8Bytes, "7836", "https://acm.hdu.edu.cn/showproblem.php?pid=7836")
	if err != nil {
		t.Fatalf("parseProblem 失败: %v", err)
	}

	if p.Title != "学位运算导致的" {
		t.Errorf("Title=%q, want 学位运算导致的", p.Title)
	}
	if p.TimeLimit != 3000 { // 6000/3000 取 Others
		t.Errorf("TimeLimit=%d, want 3000", p.TimeLimit)
	}
	if p.MemoryLimit != 512 { // 524288K / 1024
		t.Errorf("MemoryLimit=%d, want 512", p.MemoryLimit)
	}
	if len(p.Samples) != 1 {
		t.Fatalf("样例数=%d, want 1", len(p.Samples))
	}
	if !strings.HasPrefix(p.Samples[0].Input, "2") || !strings.Contains(p.Samples[0].Input, "5 6 15") {
		t.Errorf("样例输入异常: %q", p.Samples[0].Input)
	}
	if !strings.Contains(p.Samples[0].Output, "6744073707654140") {
		t.Errorf("样例输出异常: %q", p.Samples[0].Output)
	}
	// 题面应保留中文 + LaTeX
	if !strings.Contains(p.Description, "优雅") || !strings.Contains(p.Description, "$") {
		t.Errorf("题面解析异常(中文/LaTeX 丢失): %.80q", p.Description)
	}
	if p.InputFormat == "" || p.OutputFormat == "" || p.Hint == "" {
		t.Errorf("字段为空: input=%d output=%d hint=%d", len(p.InputFormat), len(p.OutputFormat), len(p.Hint))
	}
}
