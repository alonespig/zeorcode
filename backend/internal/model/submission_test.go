package model

import (
	"encoding/json"
	"testing"
)

func TestSubmissionJSONExposesOnlyPublicID(t *testing.T) {
	payload, err := json.Marshal(Submission{ID: 7, PublicID: 12345678})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := fields["id"]; got != float64(12345678) {
		t.Fatalf("public id = %v, want 12345678", got)
	}
	if _, exists := fields["ID"]; exists {
		t.Fatalf("internal primary key exposed in JSON: %s", payload)
	}
}
