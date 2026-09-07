package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/common/publicid"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"
)

type fakeSubmissionStore struct {
	created           *model.Submission
	createDispatch    repository.SubmissionDispatch
	createErr         error
	markedPublished   []repository.SubmissionDispatch
	markPublishedErr  error
	submission        *model.Submission
	submissionErr     error
	caseResults       []model.JudgeResult
	caseResultsErr    error
	requestedPublicID int64
	caseSubmissionID  int64
}

func (*fakeSubmissionStore) PublicIDExists(context.Context, int64) (bool, error) {
	return false, nil
}

func (f *fakeSubmissionStore) CreatePending(_ context.Context, submission *model.Submission) (repository.SubmissionDispatch, error) {
	copy := *submission
	f.created = &copy
	return f.createDispatch, f.createErr
}

func (*fakeSubmissionStore) GetRejudgeTargets(context.Context, repository.RejudgeScope) ([]repository.RejudgeTarget, error) {
	panic("unexpected call")
}

func (*fakeSubmissionStore) ResetForRejudge(context.Context, []int64) ([]repository.SubmissionDispatch, error) {
	panic("unexpected call")
}

func (f *fakeSubmissionStore) MarkDispatchPublished(_ context.Context, dispatch repository.SubmissionDispatch) error {
	f.markedPublished = append(f.markedPublished, dispatch)
	return f.markPublishedErr
}

func (*fakeSubmissionStore) ListWithInfo(context.Context, int, int, *int, string, int64, int64, bool) ([]repository.SubmissionListItem, int64, error) {
	panic("unexpected call")
}

func (*fakeSubmissionStore) ResolveIDByPublicID(context.Context, int64) (int64, error) {
	panic("unexpected call")
}

func (f *fakeSubmissionStore) GetSubmissionByPublicID(_ context.Context, publicID int64) (*model.Submission, error) {
	f.requestedPublicID = publicID
	return f.submission, f.submissionErr
}

func (f *fakeSubmissionStore) GetCaseResultBySubID(_ context.Context, submissionID int64) ([]model.JudgeResult, error) {
	f.caseSubmissionID = submissionID
	return f.caseResults, f.caseResultsErr
}

func (*fakeSubmissionStore) GetUserRecentAcceptedSubmissions(context.Context, int64, time.Time) ([]model.Submission, error) {
	panic("unexpected call")
}

type fakeSubmissionProblemReader struct {
	problem *model.Problem
	err     error
}

func (f *fakeSubmissionProblemReader) GetByDisplayID(context.Context, string) (*model.Problem, error) {
	return f.problem, f.err
}

func (*fakeSubmissionProblemReader) ResolveID(context.Context, string) (int64, error) {
	panic("unexpected call")
}

func (f *fakeSubmissionProblemReader) GetByID(context.Context, int64) (*model.Problem, error) {
	return f.problem, f.err
}

type fakeSubmissionUserReader struct {
	user *model.User
	err  error
}

func (*fakeSubmissionUserReader) ResolveIDByUID(context.Context, int64) (int64, error) {
	panic("unexpected call")
}

func (f *fakeSubmissionUserReader) GetUserByID(context.Context, int64) (*model.User, error) {
	return f.user, f.err
}

type fakeSubmissionCache struct {
	allowed bool
	err     error
}

func (f *fakeSubmissionCache) SetNX(context.Context, string, any, time.Duration) (bool, error) {
	return f.allowed, f.err
}

func (*fakeSubmissionCache) Delete(context.Context, ...string) error     { return nil }
func (*fakeSubmissionCache) Incr(context.Context, string) (int64, error) { return 0, nil }

type fakeProblemTestDataReader struct {
	ready bool
	err   error
	calls int
}

func (f *fakeProblemTestDataReader) HasTestData(context.Context, int64) (bool, error) {
	f.calls++
	return f.ready, f.err
}

type fakeSubmissionQueue struct {
	err      error
	enqueued []repository.SubmissionDispatch
}

type fakeLanguageResolver struct{}

func (fakeLanguageResolver) ResolveEnabledName(_ context.Context, id int) (string, error) {
	switch id {
	case 1:
		return "c++", nil
	case 2:
		return "java", nil
	case 3:
		return "python", nil
	default:
		return "", errcode.ErrUnsupportedLanguage
	}
}

func (f *fakeSubmissionQueue) EnqueueSubmission(_ context.Context, submissionID int64, version int) error {
	f.enqueued = append(f.enqueued, repository.SubmissionDispatch{SubmissionID: submissionID, Version: version})
	return f.err
}

func TestCreateSubmissionPersistsAndEnqueues(t *testing.T) {
	dispatch := repository.SubmissionDispatch{SubmissionID: 91, Version: 0}
	store := &fakeSubmissionStore{createDispatch: dispatch}
	queue := &fakeSubmissionQueue{}
	svc := NewSubmissionService(
		store,
		&fakeSubmissionUserReader{},
		&fakeSubmissionProblemReader{problem: &model.Problem{ID: 7}},
		&fakeProblemTestDataReader{ready: true},
		&fakeSubmissionCache{allowed: true},
		queue,
		fakeLanguageResolver{},
	)

	id, err := svc.CreateSubmission(context.Background(), &dto.SubmitCodeReq{
		ProblemID: "P1000",
		Language:  1,
		Code:      "int main() {}",
	}, 11, false)

	if err != nil {
		t.Fatalf("CreateSubmission() error = %v", err)
	}
	if store.created == nil {
		t.Fatal("submission was not persisted")
	}
	if !publicid.Valid(id) || id != store.created.PublicID {
		t.Fatalf("public submission ID = %d, persisted = %d", id, store.created.PublicID)
	}
	if store.created.ProblemID != 7 || store.created.UserID != 11 {
		t.Fatalf("persisted submission = %+v", store.created)
	}
	if store.created.Language != "c++" || store.created.Status != judge.Pending {
		t.Fatalf("persisted language/status = %q/%d", store.created.Language, store.created.Status)
	}
	if len(queue.enqueued) != 1 || queue.enqueued[0] != dispatch {
		t.Fatalf("enqueued dispatches = %v, want [%v]", queue.enqueued, dispatch)
	}
	if len(store.markedPublished) != 1 || store.markedPublished[0] != dispatch {
		t.Fatalf("published dispatches = %v, want [%v]", store.markedPublished, dispatch)
	}
}

func TestCreateSubmissionReturnsErrorWhenQueueFailsAfterPersist(t *testing.T) {
	queueErr := errors.New("redis unavailable")
	store := &fakeSubmissionStore{createDispatch: repository.SubmissionDispatch{SubmissionID: 91}}
	queue := &fakeSubmissionQueue{err: queueErr}
	svc := NewSubmissionService(
		store,
		&fakeSubmissionUserReader{},
		&fakeSubmissionProblemReader{problem: &model.Problem{ID: 7}},
		&fakeProblemTestDataReader{ready: true},
		&fakeSubmissionCache{allowed: true},
		queue,
		fakeLanguageResolver{},
	)

	_, err := svc.CreateSubmission(context.Background(), &dto.SubmitCodeReq{
		ProblemID: "P1000",
		Language:  3,
		Code:      "print(1)",
	}, 11, false)

	if !errors.Is(err, queueErr) {
		t.Fatalf("error = %v, want wrapped queue error", err)
	}
	if store.created == nil {
		t.Fatal("current contract persists the submission before enqueueing")
	}
	if len(queue.enqueued) != 1 || queue.enqueued[0] != (repository.SubmissionDispatch{SubmissionID: 91}) {
		t.Fatalf("enqueued dispatches = %v, want [{91 0}]", queue.enqueued)
	}
	if len(store.markedPublished) != 0 {
		t.Fatalf("published dispatches = %v, want none after queue failure", store.markedPublished)
	}
}

func TestCreateSubmissionRejectsLocalProblemWithoutTestData(t *testing.T) {
	store := &fakeSubmissionStore{}
	queue := &fakeSubmissionQueue{}
	testData := &fakeProblemTestDataReader{ready: false}
	svc := NewSubmissionService(
		store,
		&fakeSubmissionUserReader{},
		&fakeSubmissionProblemReader{problem: &model.Problem{ID: 7}},
		testData,
		&fakeSubmissionCache{allowed: true},
		queue,
		fakeLanguageResolver{},
	)

	_, err := svc.CreateSubmission(context.Background(), &dto.SubmitCodeReq{
		ProblemID: "P1000",
		Language:  1,
		Code:      "int main() {}",
	}, 11, false)

	var appErr *errcode.AppErr
	if !errors.As(err, &appErr) || appErr.Code != errcode.ProblemTestDataMissing {
		t.Fatalf("error = %v, want ProblemTestDataMissing", err)
	}
	if testData.calls != 1 {
		t.Fatalf("test data checks = %d, want 1", testData.calls)
	}
	if store.created != nil || len(queue.enqueued) != 0 {
		t.Fatalf("submission persisted/enqueued = %v/%v, want neither", store.created, queue.enqueued)
	}
}

func TestCreateSubmissionAllowsRemoteProblemWithoutLocalTestData(t *testing.T) {
	dispatch := repository.SubmissionDispatch{SubmissionID: 92}
	store := &fakeSubmissionStore{createDispatch: dispatch}
	testData := &fakeProblemTestDataReader{ready: false}
	svc := NewSubmissionService(
		store,
		&fakeSubmissionUserReader{},
		&fakeSubmissionProblemReader{problem: &model.Problem{ID: 8, OJ: "LibreOJ"}},
		testData,
		&fakeSubmissionCache{allowed: true},
		&fakeSubmissionQueue{},
		fakeLanguageResolver{},
	)

	if _, err := svc.CreateSubmission(context.Background(), &dto.SubmitCodeReq{
		ProblemID: "P2000",
		Language:  3,
		Code:      "print(1)",
	}, 11, false); err != nil {
		t.Fatalf("CreateSubmission() error = %v", err)
	}
	if testData.calls != 0 {
		t.Fatalf("local test data checks = %d, want 0 for remote problem", testData.calls)
	}
}

func TestGetSubmissionByPublicIDOnlyOwnerOrAdminSeesCodeAndCompileOutput(t *testing.T) {
	tests := []struct {
		name              string
		requesterID       int64
		isAdmin           bool
		wantCode          string
		wantCompileOutput string
	}{
		{name: "提交者可查看源码和编译输出", requesterID: 11, wantCode: "secret code", wantCompileOutput: "compiler output"},
		{name: "管理员可查看源码和编译输出", requesterID: 22, isAdmin: true, wantCode: "secret code", wantCompileOutput: "compiler output"},
		{name: "其他用户看不到源码", requesterID: 22},
		{name: "匿名用户看不到源码"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeSubmissionStore{submission: &model.Submission{
				ID: 1, PublicID: 12345678, UserID: 11, ProblemID: 7,
				Code: "secret code", Language: "c++", CompileOutput: "compiler output",
			}}
			svc := NewSubmissionService(
				store,
				&fakeSubmissionUserReader{user: &model.User{ID: 11, UID: 10000011, Username: "owner"}},
				&fakeSubmissionProblemReader{problem: &model.Problem{ID: 7, DisplayID: "P1000"}},
				&fakeProblemTestDataReader{},
				&fakeSubmissionCache{},
				&fakeSubmissionQueue{},
				fakeLanguageResolver{},
			)

			resp, err := svc.GetSubmissionByPublicID(context.Background(), 12345678, tt.requesterID, tt.isAdmin)
			if err != nil {
				t.Fatalf("GetSubmissionByPublicID() error = %v", err)
			}
			if resp.Submission.Code != tt.wantCode {
				t.Fatalf("submission code = %q, want %q", resp.Submission.Code, tt.wantCode)
			}
			if resp.Submission.CompileOutput != tt.wantCompileOutput {
				t.Fatalf("compile output = %q, want %q", resp.Submission.CompileOutput, tt.wantCompileOutput)
			}
			if resp.Submission.ID != 12345678 || store.requestedPublicID != 12345678 {
				t.Fatalf("public ID response/request = %d/%d", resp.Submission.ID, store.requestedPublicID)
			}
			if store.caseSubmissionID != 1 {
				t.Fatalf("case results queried with ID %d, want internal ID 1", store.caseSubmissionID)
			}
		})
	}
}

func TestGetSubmissionByPublicIDRejectsOtherUsersContestAndHomeworkSubmissions(t *testing.T) {
	tests := []struct {
		name       string
		contestID  int64
		homeworkID int64
	}{
		{name: "比赛提交", contestID: 9},
		{name: "作业提交", homeworkID: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeSubmissionStore{submission: &model.Submission{
				ID: 1, UserID: 11, ProblemID: 7, ContestID: tt.contestID, HomeworkID: tt.homeworkID,
			}}
			svc := NewSubmissionService(
				store,
				&fakeSubmissionUserReader{},
				&fakeSubmissionProblemReader{},
				&fakeProblemTestDataReader{},
				&fakeSubmissionCache{},
				&fakeSubmissionQueue{},
				fakeLanguageResolver{},
			)

			_, err := svc.GetSubmissionByPublicID(context.Background(), 12345678, 22, false)
			var appErr *errcode.AppErr
			if !errors.As(err, &appErr) || appErr.Code != errcode.SubmissionNotFound {
				t.Fatalf("GetSubmissionByPublicID() error = %v, want SubmissionNotFound", err)
			}
		})
	}
}
