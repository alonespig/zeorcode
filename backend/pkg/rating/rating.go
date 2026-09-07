// Package rating 实现 Codeforces 经典竞赛 rating 算法（纯函数、无副作用、可离线单测）。
//
// 算法：期望名次 seed → 几何均值目标 √(rank·seed) → 二分解目标分 →
// delta=(目标−当前)/2 → 两步归一化（整体零和微调 + 高分段稳定修正）。
package rating

import (
	"math"
	"sort"
)

// Player 一名参赛者的输入：当前分 + 最终名次（并列共享同一名次，rank 从 1 起）。
type Player struct {
	Rating int
	Rank   int
}

// Result 结算结果，与传入的 players 一一对应（同下标）。
type Result struct {
	Delta     int
	NewRating int
}

// eloWinProbability 评分 ra 的选手战胜 rb 的概率。
func eloWinProbability(ra, rb float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (rb-ra)/400.0))
}

// Calculate 计算每人的 rating 变化。输入顺序即输出顺序。
// 少于 2 人无法有意义地计分，返回全 0（分数不变）。
func Calculate(players []Player) []Result {
	n := len(players)
	res := make([]Result, n)
	for i := range players {
		res[i] = Result{Delta: 0, NewRating: players[i].Rating}
	}
	if n < 2 {
		return res
	}

	ratings := make([]float64, n)
	for i, p := range players {
		ratings[i] = float64(p.Rating)
	}

	// seed_i = 1 + Σ_{j≠i} P(j 胜 i)：把 i 当作当前分时的期望名次。
	seed := make([]float64, n)
	for i := 0; i < n; i++ {
		s := 1.0
		for j := 0; j < n; j++ {
			if j != i {
				s += eloWinProbability(ratings[j], ratings[i])
			}
		}
		seed[i] = s
	}

	// 目标分：解 R' 使 seed_i(R') = √(rank_i · seed_i)；delta = (R'−R)/2。
	delta := make([]int, n)
	for i := 0; i < n; i++ {
		need := math.Sqrt(float64(players[i].Rank) * seed[i])
		target := ratingForSeed(ratings, i, need)
		delta[i] = (target - players[i].Rating) / 2
	}

	// —— 两步归一化 ——
	// ① 整体零和微调（略偏负，抑制通胀）
	sum := 0
	for _, d := range delta {
		sum += d
	}
	inc1 := int(math.Round(float64(-sum)/float64(n))) - 1
	for i := range delta {
		delta[i] += inc1
	}

	// ② 高分段稳定修正：按原分降序取前 s 人，把其 delta 之和拉回 [-10,0]
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return players[order[a]].Rating > players[order[b]].Rating
	})
	s := int(math.Round(4 * math.Sqrt(float64(n))))
	if s > n {
		s = n
	}
	if s < 1 {
		s = 1
	}
	sumTop := 0
	for k := 0; k < s; k++ {
		sumTop += delta[order[k]]
	}
	inc2 := int(math.Max(math.Min(float64(-sumTop)/float64(s), 0), -10))
	for i := range delta {
		delta[i] += inc2
	}

	for i := 0; i < n; i++ {
		res[i] = Result{Delta: delta[i], NewRating: players[i].Rating + delta[i]}
	}
	return res
}

// ratingForSeed 二分找分数 r，使「self 若为 r 分时的 seed」≈ want。
// seed 对 r 单调递减，故可二分（范围 [1, 8000]）。
func ratingForSeed(ratings []float64, self int, want float64) int {
	lo, hi := 1, 8000
	for hi-lo > 1 {
		mid := (lo + hi) / 2
		s := 1.0
		for j := 0; j < len(ratings); j++ {
			if j != self {
				s += eloWinProbability(ratings[j], float64(mid))
			}
		}
		if s < want {
			hi = mid
		} else {
			lo = mid
		}
	}
	return lo
}
