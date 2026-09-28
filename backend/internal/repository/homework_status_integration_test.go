//go:build integration

package repository

import (
	"reflect"
	"testing"

	"zoj/internal/model"
	"zoj/pkg/judge"
)

func TestHomeworkUserProblemStatusesExcludeOtherContexts(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	subs := []model.Submission{
		{UserID: 1, ProblemID: 1, Status: judge.Accepted}, // 仅题库通过
		{UserID: 1, ProblemID: 2, ContestID: 5, Status: judge.Accepted},
		{UserID: 1, ProblemID: 3, HomeworkID: 11, Status: judge.Accepted},
		{UserID: 2, ProblemID: 4, HomeworkID: 10, Status: judge.Accepted},
		{UserID: 1, ProblemID: 5, HomeworkID: 10, Status: judge.WrongAnswer},
		{UserID: 1, ProblemID: 6, HomeworkID: 10, Status: judge.Pending},
		{UserID: 1, ProblemID: 7, HomeworkID: 10, Status: judge.Accepted},
		{UserID: 1, ProblemID: 7, HomeworkID: 10, Status: judge.WrongAnswer}, // AC 不被后续 WA 覆盖
		{UserID: 1, ProblemID: 8, HomeworkID: 10, Status: judge.WrongAnswer},
		{UserID: 1, ProblemID: 8, HomeworkID: 10, Status: judge.CompileError},
		{UserID: 1, ProblemID: 9, HomeworkID: 10, Status: judge.Accepted}, // 不在请求题目集合中
	}
	for i := range subs {
		subs[i].PublicID = int64(i + 1)
	}
	if err := db.Create(&subs).Error; err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetHomeworkUserProblemStatuses(t.Context(), 10, 1, []int64{1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int64]int{5: judge.WrongAnswer, 6: judge.Pending, 7: judge.Accepted, 8: judge.CompileError}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statuses = %v, want %v", got, want)
	}
}
