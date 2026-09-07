package service

import (
	"testing"
	"time"

	"zoj/internal/model"
	"zoj/internal/repository"
)

func TestBuildHomeworkRankRowsIncludesMembersWithoutSubmissions(t *testing.T) {
	refs := []repository.HomeworkProblemRef{
		{ProblemID: 10, Sort: 0},
		{ProblemID: 20, Sort: 1},
	}
	displayByID := map[int64]string{10: "A", 20: "B"}
	members := []model.TeamMember{
		{UserID: 1},
		{UserID: 2},
		{UserID: 3},
	}
	studentNo1 := "20260001"
	studentNo2 := "20269999"
	studentNo3 := "20260000" // 更小，用来证明 0 分时有提交者仍排在未提交者前面
	users := []model.User{
		{ID: 1, UID: 10000001, Username: "alice", StudentNo: &studentNo1, RealName: "张一"},
		{ID: 2, UID: 10000002, Username: "bob", StudentNo: &studentNo2, RealName: "张二"},
		{ID: 3, UID: 10000003, Username: "carol", StudentNo: &studentNo3, RealName: "张三"},
	}
	best := []repository.HomeworkBestScore{
		{UserID: 1, ProblemID: 10, Best: 100, AchievedAt: time.Date(2026, 8, 26, 8, 0, 0, 0, time.Local)},
		{UserID: 2, ProblemID: 10, Best: 0, AchievedAt: time.Date(2026, 8, 26, 9, 0, 0, 0, time.Local)},
		// 已退队用户的历史成绩不能出现在榜单里。
		{UserID: 99, ProblemID: 10, Best: 100, AchievedAt: time.Date(2026, 8, 26, 7, 0, 0, 0, time.Local)},
	}

	rows := buildHomeworkRankRows(refs, displayByID, members, users, best)
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].UID != 10000001 || rows[0].TotalScore != 100 || rows[0].SolvedCount != 1 {
		t.Fatalf("first row = %+v", rows[0])
	}
	if rows[1].UID != 10000002 || rows[1].TotalScore != 0 {
		t.Fatalf("second row = %+v; submitted zero-score member should precede non-submitter", rows[1])
	}
	if rows[2].UID != 10000003 || rows[2].TotalScore != 0 || rows[2].SolvedCount != 0 {
		t.Fatalf("non-submitter row = %+v", rows[2])
	}
	if len(rows[2].Cells) != 2 || rows[2].Cells[0].Score != 0 || rows[2].Cells[1].Score != 0 {
		t.Fatalf("non-submitter cells = %+v, want two zero-score cells", rows[2].Cells)
	}
	for i, row := range rows {
		if row.Rank != i+1 {
			t.Errorf("row %d rank = %d, want %d", i, row.Rank, i+1)
		}
	}
}

func TestBuildHomeworkRankRowsWithNoScoresStillListsAllMembers(t *testing.T) {
	studentNo1 := "20260002"
	studentNo2 := "20260001"
	rows := buildHomeworkRankRows(
		[]repository.HomeworkProblemRef{{ProblemID: 10}},
		map[int64]string{10: "A"},
		[]model.TeamMember{{UserID: 1}, {UserID: 2}},
		[]model.User{
			{ID: 1, UID: 10000001, StudentNo: &studentNo1},
			{ID: 2, UID: 10000002, StudentNo: &studentNo2},
		},
		nil,
	)
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].UID != 10000002 || rows[1].UID != 10000001 {
		t.Fatalf("zero-score order = [%d, %d], want student number order", rows[0].UID, rows[1].UID)
	}
}
