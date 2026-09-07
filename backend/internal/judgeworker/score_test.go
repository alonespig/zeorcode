package judgeworker

import (
	"testing"

	"zoj/internal/model"
	"zoj/pkg/judge"
)

func results(statuses ...int) []*model.JudgeResult {
	out := make([]*model.JudgeResult, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, &model.JudgeResult{Status: s})
	}
	return out
}

func TestCaseAverageScore(t *testing.T) {
	tests := []struct {
		name    string
		results []*model.JudgeResult
		full    int
		want    int
	}{
		{
			name:    "全部通过给满分",
			results: results(judge.Accepted, judge.Accepted, judge.Accepted),
			full:    100,
			want:    100,
		},
		{
			name:    "全错给0",
			results: results(judge.WrongAnswer, judge.WrongAnswer),
			full:    100,
			want:    0,
		},
		{
			name:    "过半数按比例给部分分",
			results: results(judge.Accepted, judge.Accepted, judge.WrongAnswer, judge.WrongAnswer),
			full:    100,
			want:    50,
		},
		{
			name:    "非整除四舍五入",
			results: results(judge.Accepted, judge.WrongAnswer, judge.WrongAnswer),
			full:    100,
			want:    33, // 1/3*100 = 33.33 -> 33
		},
		{
			name:    "TLE/RE 等非 AC 一律不计通过",
			results: results(judge.Accepted, judge.TimeLimitExceeded, judge.RuntimeError),
			full:    100,
			want:    33,
		},
		{
			name:    "尊重比赛自定义满分",
			results: results(judge.Accepted, judge.WrongAnswer),
			full:    500,
			want:    250,
		},
		{
			name:    "没有测试点结果给0，不能除零",
			results: nil,
			full:    100,
			want:    0,
		},
		{
			name:    "满分非正给0",
			results: results(judge.Accepted),
			full:    0,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := caseAverageScore(tt.results, tt.full); got != tt.want {
				t.Fatalf("caseAverageScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestSubmissionScoreRouting 覆盖「作业提交也要算分」这个本次新增的关键行为：
// 改动前 Score 只在 ContestID != 0 时计算，作业提交分数会恒为 0。
func TestSubmissionScoreRouting(t *testing.T) {
	w := &Worker{}
	half := results(judge.Accepted, judge.WrongAnswer)

	t.Run("作业提交按满分100算部分分", func(t *testing.T) {
		s := &model.Submission{HomeworkID: 7}
		if got := w.submissionScore(s, half); got != 50 {
			t.Fatalf("homework score = %d, want 50", got)
		}
	})

	t.Run("普通练习不计分", func(t *testing.T) {
		s := &model.Submission{}
		if got := w.submissionScore(s, half); got != 0 {
			t.Fatalf("practice score = %d, want 0", got)
		}
	})
}
