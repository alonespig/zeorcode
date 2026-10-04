package judge

import (
	"encoding/json"
	"testing"
)

func TestSubmissionEventInfoJSON(t *testing.T) {
	ev := SubmissionEventInfo{
		ID: 123, Language: "c++", Status: 1, TimeUsed: 100, MemoryUsed: 2048,
		CreatedAt: "2026-10-05 12:34:56",
	}
	got, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":123,"language":"c++","status":1,"time":100,"memory":2048,"createdAt":"2026-10-05 12:34:56"}`
	if string(got) != want {
		t.Fatalf("event JSON = %s\nwant %s", got, want)
	}
}

func TestSubmissionCaseResultJSON(t *testing.T) {
	c := SubmissionCaseResult{ID: 1, Status: 2, TimeUsed: 10, MemoryUsed: 1024}
	got, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":1,"status":2,"time":10,"memory":1024}`
	if string(got) != want {
		t.Fatalf("case JSON = %s\nwant %s", got, want)
	}
}

func TestEmptyCaseResultsJSON(t *testing.T) {
	got, err := json.Marshal([]SubmissionCaseResult{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "[]" {
		t.Fatalf("empty case results = %s, want []", got)
	}
}
