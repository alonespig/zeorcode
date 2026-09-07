package nowcoder

import (
	"testing"

	"zoj/pkg/judge"
)

func TestExtractIDs(t *testing.T) {
	// 覆盖 JSON 风格 和 JS 对象风格；并确保 subTagId 不会被当成 tagId
	cases := []string{
		`window.pageConfig={"questionId":11714784,"tagId":4,"subTagId":0}`,
		`var x = { questionId: 11714784, subTagId: 0, tagId: 4 }`,
	}
	for _, html := range cases {
		qid := reQuestionID.FindStringSubmatch(html)
		tid := reTagID.FindStringSubmatch(html)
		if qid == nil || qid[1] != "11714784" {
			t.Fatalf("questionId 抽取失败: %v (输入 %q)", qid, html)
		}
		if tid == nil || tid[1] != "4" {
			t.Fatalf("tagId 抽取失败: %v (输入 %q)", tid, html)
		}
	}
}

func TestMapStatus(t *testing.T) {
	cases := []struct {
		status int
		desc   string
		want   int
	}{
		{12, "编译错误", judge.CompileError}, // 用户实测确认
		{0, "答案正确", judge.Accepted},
		{0, "答案错误", judge.WrongAnswer},
		{0, "程序运行超时", judge.TimeLimitExceeded},
		{0, "运行内存超限", judge.MemoryLimitExceeded},
		{0, "程序运行错误", judge.RuntimeError},
		{0, "正在判题", judge.Pending},
		{5, "", judge.Accepted}, // 仅靠状态码兜底
	}
	for _, c := range cases {
		if got := mapStatus(c.status, c.desc); got != c.want {
			t.Errorf("mapStatus(%d,%q)=%d, want %d", c.status, c.desc, got, c.want)
		}
	}
}
