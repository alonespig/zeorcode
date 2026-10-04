package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSubmissionEventInfoDoesNotExposeCode(t *testing.T) {
	payload, err := json.Marshal(SubmissionEventInfo{ID: 1, Language: "c++", Status: 1})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(payload), `"code"`) {
		t.Fatalf("event payload exposes source code field: %s", payload)
	}
}
