package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"zoj/internal/model"
	judgeapi "zoj/pkg/judge"
)

type fakeCodeRunPool struct {
	client     *judgeapi.Client
	markedDown int
}

func (f *fakeCodeRunPool) Size() int {
	if f.client == nil {
		return 0
	}
	return 1
}

func (f *fakeCodeRunPool) Next() *judgeapi.Client { return f.client }

func (f *fakeCodeRunPool) MarkDown(*judgeapi.Client) { f.markedDown++ }

func TestCodeRunServiceRunsCustomInputWithoutCreatingSubmission(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/run":
			requestCount++
			w.Header().Set("Content-Type", "application/json")
			if requestCount == 1 {
				_, _ = fmt.Fprint(w, `[{"status":"Accepted","files":{"stdout":"","stderr":""},"fileIds":{"a":"artifact-1"}}]`)
				return
			}
			_, _ = fmt.Fprint(w, `[{"status":"Accepted","time":2000000,"runTime":2000000,"memory":4096,"files":{"stdout":"3\n","stderr":""},"fileIds":{}}]`)
		case r.Method == http.MethodDelete && r.URL.Path == "/file/artifact-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected judge request: %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	pool := &fakeCodeRunPool{client: judgeapi.NewClient(server.URL)}
	svc := NewCodeRunService(
		&fakeSubmissionProblemReader{problem: &model.Problem{ID: 7, DisplayID: "L101", TimeLimit: 2000, MemoryLimit: 256}},
		fakeLanguageResolver{},
		&fakeSubmissionCache{allowed: true},
		pool,
	)

	result, err := svc.Run(context.Background(), CodeRunParams{
		ProblemID: "L101",
		Language:  1,
		Code:      "int main() {}",
		Input:     "1 2\n",
	}, 42, false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != judgeapi.Accepted || result.Stdout != "3\n" {
		t.Fatalf("Run() result = %#v", result)
	}
	if result.TimeUsed != 2 || result.MemoryUsed != 4 {
		t.Fatalf("resource usage = %d ms, %d KB", result.TimeUsed, result.MemoryUsed)
	}
	if pool.markedDown != 0 {
		t.Fatalf("healthy judge marked down %d times", pool.markedDown)
	}
}

func TestValidateCodeRunParamsRejectsOversizedInput(t *testing.T) {
	params := CodeRunParams{
		ProblemID: "L101",
		Language:  1,
		Code:      "int main() {}",
		Input:     string(make([]byte, maxCodeRunInputBytes+1)),
	}
	if err := validateCodeRunParams(params); err == nil {
		t.Fatal("validateCodeRunParams() error = nil, want oversized input error")
	}
}
