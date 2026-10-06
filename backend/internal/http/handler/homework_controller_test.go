package handler

import "testing"

func TestSafeSpreadsheetCell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "普通文本", input: "20260001", want: "20260001"},
		{name: "公式", input: "=1+1", want: "'=1+1"},
		{name: "前导空格公式", input: "  +SUM(A1:A2)", want: "'  +SUM(A1:A2)"},
		{name: "邮箱前缀", input: "@cmd", want: "'@cmd"},
		{name: "空白", input: "  ", want: "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeSpreadsheetCell(tt.input); got != tt.want {
				t.Fatalf("safeSpreadsheetCell(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHomeworkRankGenderLabel(t *testing.T) {
	tests := []struct {
		gender int
		want   string
	}{
		{gender: 1, want: "男"},
		{gender: 2, want: "女"},
		{gender: 0, want: "未设置"},
		{gender: 3, want: "未设置"},
	}

	for _, tt := range tests {
		if got := homeworkRankGenderLabel(tt.gender); got != tt.want {
			t.Fatalf("homeworkRankGenderLabel(%d) = %q, want %q", tt.gender, got, tt.want)
		}
	}
}
