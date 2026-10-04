package service

import (
	"testing"

	"zoj/pkg/errcode"
)

func TestParseHomeworkWindow(t *testing.T) {
	tests := []struct {
		name     string
		req      SaveHomeworkParams
		wantCode errcode.Code // 0 表示放行
	}{
		{
			name: "正常时间窗",
			req:  SaveHomeworkParams{StartTime: "2026-08-01 00:00:00", EndTime: "2026-09-01 00:00:00"},
		},
		{
			name:     "结束早于开始",
			req:      SaveHomeworkParams{StartTime: "2026-09-01 00:00:00", EndTime: "2026-08-01 00:00:00"},
			wantCode: errcode.InvalidParams,
		},
		{
			name:     "结束等于开始也算非法",
			req:      SaveHomeworkParams{StartTime: "2026-08-01 00:00:00", EndTime: "2026-08-01 00:00:00"},
			wantCode: errcode.InvalidParams,
		},
		{
			name:     "开始时间格式错",
			req:      SaveHomeworkParams{StartTime: "2026/08/01", EndTime: "2026-09-01 00:00:00"},
			wantCode: errcode.InvalidParams,
		},
		{
			name:     "结束时间格式错",
			req:      SaveHomeworkParams{StartTime: "2026-08-01 00:00:00", EndTime: "bad"},
			wantCode: errcode.InvalidParams,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseHomeworkWindow(tt.req)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("parseHomeworkWindow() error = %v, want nil", err)
				}
				if !end.After(start) {
					t.Fatalf("end %v should be after start %v", end, start)
				}
				return
			}
			requireAppErrorCode(t, err, tt.wantCode)
		})
	}
}
