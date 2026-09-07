package mq

import (
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestToTask(t *testing.T) {
	tests := []struct {
		name string
		msg  goredis.XMessage
		want *StreamTask
	}{
		{
			name: "versioned message",
			msg: goredis.XMessage{ID: "1-0", Values: map[string]any{
				submissionField: "42",
				versionField:    "3",
			}},
			want: &StreamTask{SubmissionID: 42, Version: 3, Versioned: true, MsgID: "1-0"},
		},
		{
			name: "legacy message is initial version",
			msg: goredis.XMessage{ID: "2-0", Values: map[string]any{
				submissionField: int64(7),
			}},
			want: &StreamTask{SubmissionID: 7, MsgID: "2-0"},
		},
		{
			name: "version zero is explicitly versioned",
			msg: goredis.XMessage{ID: "3-0", Values: map[string]any{
				submissionField: 9,
				versionField:    0,
			}},
			want: &StreamTask{SubmissionID: 9, Versioned: true, MsgID: "3-0"},
		},
		{
			name: "missing submission id",
			msg:  goredis.XMessage{ID: "4-0", Values: map[string]any{versionField: 1}},
		},
		{
			name: "invalid submission id",
			msg:  goredis.XMessage{ID: "5-0", Values: map[string]any{submissionField: 0}},
		},
		{
			name: "negative version",
			msg: goredis.XMessage{ID: "6-0", Values: map[string]any{
				submissionField: 1,
				versionField:    -1,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTask(tt.msg)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("toTask() = %#v, want nil", got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("toTask() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
