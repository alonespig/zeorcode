// Package scoring 实现 Codeforces 动态分（题目满分随时间衰减 + 错交扣分，纯函数）。
package scoring

import "math"

// CFScore 计算一道题的动态得分：
//
//	score = max(0.3·x, x − floor(120·x·t/(250·d)) − 50·w)
//
// x=题目初始分, t=首次 AC 时刻(距开始的分钟), d=比赛时长(分钟),
// w=首次 AC 前的错误提交数(不含 CE)。未 AC(accepted=false) 返回 0。
// 只要 AC，最低不低于初始分的 30%。
func CFScore(x, t, d, w int, accepted bool) int {
	if !accepted || x <= 0 {
		return 0
	}
	if d <= 0 {
		d = 1
	}
	if t < 0 {
		t = 0
	} else if t > d {
		t = d
	}
	if w < 0 {
		w = 0
	}

	decay := int(math.Floor(float64(120*x*t) / float64(250*d)))
	raw := x - decay - 50*w

	floor := int(math.Round(0.3 * float64(x)))
	if raw < floor {
		return floor
	}
	return raw
}
