package judge

import (
	"testing"

	judgeapi "zoj/pkg/judge"
)

// TestJudgeOutput 锁定 AC / PE / WA 判定顺序与空白容忍规则：
// 先去行末空白(rtrim)比对判 AC，再去全部空白(strip)比对判 PE，否则 WA。
func TestJudgeOutput(t *testing.T) {
	expected := "3\n"
	acHash := judgeapi.MD5(judgeapi.RtrimOutput(expected))
	peHash := judgeapi.MD5(judgeapi.StripAllSpace(expected))

	tests := []struct {
		name   string
		stdout string
		want   int
	}{
		{"ac exact", "3\n", judgeapi.Accepted},
		{"ac trailing space", "3 \n", judgeapi.Accepted},
		{"ac trailing newlines", "3\n\n\n", judgeapi.Accepted},
		{"pe leading space", " 3\n", judgeapi.PresentationError},
		{"pe leading and trailing space", " 3 \n", judgeapi.PresentationError},
		{"wa different value", "4\n", judgeapi.WrongAnswer},
		{"wa extra content", "3 4\n", judgeapi.WrongAnswer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := judgeOutput(tt.stdout, acHash, peHash); got != tt.want {
				t.Fatalf("judgeOutput(%q) = %d, want %d", tt.stdout, got, tt.want)
			}
		})
	}
}
