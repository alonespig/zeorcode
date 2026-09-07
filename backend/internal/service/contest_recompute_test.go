package service

import (
	"testing"
	"time"

	"zoj/internal/common/consts"
	"zoj/internal/repository"
	"zoj/pkg/judge"
)

func TestContestProblemLabel(t *testing.T) {
	tests := []struct {
		index int
		want  string
	}{
		{index: 0, want: "A"},
		{index: 25, want: "Z"},
		{index: 26, want: "AA"},
		{index: 27, want: "AB"},
		{index: 51, want: "AZ"},
		{index: 52, want: "BA"},
	}

	for _, tt := range tests {
		if got := contestProblemLabel(tt.index); got != tt.want {
			t.Fatalf("contestProblemLabel(%d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}

func t(sec int) time.Time { return time.Date(2026, 1, 1, 0, 0, sec, 0, time.UTC) }

func row(status, score, sec int) repository.RecomputeSubmissionRow {
	return repository.RecomputeSubmissionRow{UserID: 1, ProblemID: 2, Status: status, Score: score, CreatedAt: t(sec)}
}

func TestBuildContestUCP_ACM(t_ *testing.T) {
	// 两次错(WA/TLE) → AC → AC 之后再一次 WA(应被冻结忽略)
	subs := []repository.RecomputeSubmissionRow{
		row(judge.WrongAnswer, 0, 0),
		row(judge.TimeLimitExceeded, 0, 10),
		row(judge.Accepted, 100, 20),
		row(judge.WrongAnswer, 0, 30),
	}
	ucp := buildContestUCP(consts.ContestACM, 9, 1, 2, subs)
	if ucp.Status != judge.Accepted {
		t_.Fatalf("status: want Accepted, got %d", ucp.Status)
	}
	if ucp.AcCount != 1 || ucp.UnAcCount != 2 || ucp.SubCount != 4 {
		t_.Fatalf("counts: ac=%d unac=%d sub=%d, want 1/2/4", ucp.AcCount, ucp.UnAcCount, ucp.SubCount)
	}
	if ucp.AcTime == nil || !ucp.AcTime.Equal(t(20)) {
		t_.Fatalf("acTime: want t=20, got %v", ucp.AcTime)
	}
}

func TestBuildContestUCP_ACM_NeverAC(t_ *testing.T) {
	subs := []repository.RecomputeSubmissionRow{
		row(judge.WrongAnswer, 0, 0),
		row(judge.TimeLimitExceeded, 0, 10),
	}
	ucp := buildContestUCP(consts.ContestACM, 9, 1, 2, subs)
	if ucp.Status != judge.TimeLimitExceeded {
		t_.Fatalf("status: want last=TLE, got %d", ucp.Status)
	}
	if ucp.AcCount != 0 || ucp.UnAcCount != 2 || ucp.AcTime != nil {
		t_.Fatalf("never-AC: ac=%d unac=%d acTime=%v", ucp.AcCount, ucp.UnAcCount, ucp.AcTime)
	}
}

func TestBuildContestUCP_OI_LastScore(t_ *testing.T) {
	// OI 取最后一次提交的分
	subs := []repository.RecomputeSubmissionRow{
		row(judge.WrongAnswer, 30, 0),
		row(judge.WrongAnswer, 80, 10),
		row(judge.WrongAnswer, 20, 20),
	}
	ucp := buildContestUCP(consts.ContestOI, 9, 1, 2, subs)
	if ucp.Score != 20 {
		t_.Fatalf("OI score: want last=20, got %d", ucp.Score)
	}
	if ucp.AcTime == nil || !ucp.AcTime.Equal(t(20)) {
		t_.Fatalf("OI acTime: want t=20, got %v", ucp.AcTime)
	}
	if ucp.SubCount != 3 {
		t_.Fatalf("OI subCount: want 3, got %d", ucp.SubCount)
	}
}

func TestBuildContestUCP_IOI_MaxScore(t_ *testing.T) {
	// IOI 取历史最高分及其达成时间
	subs := []repository.RecomputeSubmissionRow{
		row(judge.WrongAnswer, 30, 0),
		row(judge.WrongAnswer, 80, 10),
		row(judge.WrongAnswer, 50, 20),
	}
	ucp := buildContestUCP(consts.ContestIOI, 9, 1, 2, subs)
	if ucp.Score != 80 {
		t_.Fatalf("IOI score: want max=80, got %d", ucp.Score)
	}
	if ucp.AcTime == nil || !ucp.AcTime.Equal(t(10)) {
		t_.Fatalf("IOI acTime: want t=10 (max achieved), got %v", ucp.AcTime)
	}
}
