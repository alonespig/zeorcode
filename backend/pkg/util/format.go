package util

import (
	"fmt"
	"time"
)

// FormatDurationHHMMSS 将两个时间差格式化为 hh:mm:ss
func FormatDurationHHMMSS(t1, t2 time.Time) string {
	d := t2.Sub(t1)

	// 处理负数情况
	if d < 0 {
		d = -d
	}

	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
