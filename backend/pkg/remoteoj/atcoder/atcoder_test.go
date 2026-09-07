package atcoder

import (
	"os"
	"testing"
)

func TestSplitPID(t *testing.T) {
	cases := []struct{ pid, contest, task string }{
		{"abc462_a", "abc462", "abc462_a"},
		{"abc462/abc462_a", "abc462", "abc462_a"},
		{"arc100_c", "arc100", "arc100_c"},
	}
	for _, c := range cases {
		gc, gt := splitPID(c.pid)
		if gc != c.contest || gt != c.task {
			t.Errorf("splitPID(%q)=(%q,%q), want (%q,%q)", c.pid, gc, gt, c.contest, c.task)
		}
	}
}

func TestParseProblem(t *testing.T) {
	htmlBytes, err := os.ReadFile("testdata/abc462_a.html")
	if err != nil {
		t.Fatalf("读样本失败: %v", err)
	}
	p, err := parseProblem(htmlBytes, "abc462_a", "https://atcoder.jp/contests/abc462/tasks/abc462_a")
	if err != nil {
		t.Fatalf("parseProblem 失败: %v", err)
	}

	if p.Title != "A - Secret Numbers" {
		t.Errorf("Title=%q", p.Title)
	}
	if p.TimeLimit != 2000 {
		t.Errorf("TimeLimit=%d, want 2000", p.TimeLimit)
	}
	if p.MemoryLimit != 1024 {
		t.Errorf("MemoryLimit=%d, want 1024", p.MemoryLimit)
	}
	if len(p.Samples) != 4 {
		t.Fatalf("样例数=%d, want 4", len(p.Samples))
	}
	if p.Samples[0].Input != "abc462" || p.Samples[0].Output != "462" {
		t.Errorf("样例1 = (%q,%q), want (abc462,462)", p.Samples[0].Input, p.Samples[0].Output)
	}
	if p.Samples[3].Input != "10plus2is12" || p.Samples[3].Output != "10212" {
		t.Errorf("样例4 = (%q,%q)", p.Samples[3].Input, p.Samples[3].Output)
	}
	if p.Description == "" || p.InputFormat == "" || p.OutputFormat == "" {
		t.Errorf("题面字段为空: desc=%d input=%d output=%d", len(p.Description), len(p.InputFormat), len(p.OutputFormat))
	}
}
