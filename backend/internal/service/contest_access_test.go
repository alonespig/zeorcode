package service

import (
	"errors"
	"testing"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/model"
)

func requireAppErrorCode(t *testing.T, err error, want errcode.Code) {
	t.Helper()
	var appErr *errcode.AppErr
	if !errors.As(err, &appErr) || appErr.Code != want {
		t.Fatalf("error = %v, want code %d", err, want)
	}
}

func TestContestProblemAccessPolicy(t *testing.T) {
	now := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		contest    model.Contest
		registered bool
		isAdmin    bool
		wantCode   errcode.Code
	}{
		{
			name:    "管理员可在开始前查看",
			contest: model.Contest{StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour)},
			isAdmin: true,
		},
		{
			name:     "普通用户不能提前查看",
			contest:  model.Contest{StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour)},
			wantCode: errcode.ContestNotStart,
		},
		{
			name:     "未报名用户不能查看进行中比赛题目",
			contest:  model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)},
			wantCode: errcode.ContestNotRegistered,
		},
		{
			name:       "已报名用户可查看进行中比赛题目",
			contest:    model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)},
			registered: true,
		},
		{
			name:       "已报名用户可查看已结束比赛题目",
			contest:    model.Contest{StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Hour)},
			registered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := contestProblemAccessPolicy(&tt.contest, tt.registered, tt.isAdmin, now)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("policy error = %v, want nil", err)
				}
				return
			}
			requireAppErrorCode(t, err, tt.wantCode)
		})
	}
}

func TestContestSubmissionPolicy(t *testing.T) {
	now := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		contest    model.Contest
		registered bool
		wantCode   errcode.Code
	}{
		{
			name:     "比赛开始前不能提交",
			contest:  model.Contest{StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour)},
			wantCode: errcode.ContestNotStart,
		},
		{
			name:     "未报名用户不能提交",
			contest:  model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)},
			wantCode: errcode.ContestNotRegistered,
		},
		{
			name:       "已报名用户可在比赛中提交",
			contest:    model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)},
			registered: true,
		},
		{
			name:       "比赛结束时不能提交",
			contest:    model.Contest{StartTime: now.Add(-time.Hour), EndTime: now},
			registered: true,
			wantCode:   errcode.ContestFinished,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := contestSubmissionPolicy(&tt.contest, tt.registered, now)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("policy error = %v, want nil", err)
				}
				return
			}
			requireAppErrorCode(t, err, tt.wantCode)
		})
	}
}

func TestContestEndTimeFallsBackToDurationMinutes(t *testing.T) {
	start := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	contest := model.Contest{StartTime: start, Duration: 90}
	want := start.Add(90 * time.Minute)
	if got := contestEndTime(contest); !got.Equal(want) {
		t.Fatalf("contestEndTime() = %v, want %v", got, want)
	}
}

func TestRatingSettleDue(t *testing.T) {
	end := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	contest := model.Contest{StartTime: end.Add(-2 * time.Hour), EndTime: end}
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "比赛进行中不结算", now: end.Add(-time.Minute), want: false},
		{name: "刚结束仍在宽限期内不结算", now: end.Add(ratingSettleGrace - time.Second), want: false},
		{name: "宽限期满可结算", now: end.Add(ratingSettleGrace), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ratingSettleDue(contest, tt.now); got != tt.want {
				t.Fatalf("ratingSettleDue() = %v, want %v", got, tt.want)
			}
		})
	}
}
