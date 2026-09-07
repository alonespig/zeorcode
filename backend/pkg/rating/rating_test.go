package rating

import "testing"

// 少于 2 人不计分。
func TestCalculate_TooFew(t *testing.T) {
	for _, players := range [][]Player{nil, {{Rating: 1500, Rank: 1}}} {
		for i, r := range Calculate(players) {
			if r.Delta != 0 || r.NewRating != players[i].Rating {
				t.Fatalf("少于2人应不变: got %+v", r)
			}
		}
	}
}

// 同分对决：名次靠前者涨、靠后者掉。
func TestCalculate_EqualRatingWinnerGains(t *testing.T) {
	res := Calculate([]Player{
		{Rating: 1500, Rank: 1},
		{Rating: 1500, Rank: 2},
	})
	if res[0].Delta <= 0 {
		t.Errorf("第一名应涨分，got delta=%d", res[0].Delta)
	}
	if res[1].Delta >= 0 {
		t.Errorf("末名应掉分，got delta=%d", res[1].Delta)
	}
}

// 名次越靠后，delta 越小（单调）。
func TestCalculate_Monotonic(t *testing.T) {
	players := make([]Player, 6)
	for i := range players {
		players[i] = Player{Rating: 1500, Rank: i + 1}
	}
	res := Calculate(players)
	for i := 1; i < len(res); i++ {
		if res[i].Delta > res[i-1].Delta {
			t.Errorf("delta 应随名次下降不增：rank%d=%d > rank%d=%d",
				i+1, res[i].Delta, i, res[i-1].Delta)
		}
	}
}

// 总变化不应为正（两步归一化抑制通胀）。
func TestCalculate_SumNotPositive(t *testing.T) {
	players := []Player{
		{Rating: 2000, Rank: 3},
		{Rating: 1500, Rank: 1},
		{Rating: 1200, Rank: 2},
		{Rating: 1800, Rank: 4},
		{Rating: 1600, Rank: 5},
	}
	sum := 0
	for _, r := range Calculate(players) {
		sum += r.Delta
	}
	if sum > 0 {
		t.Errorf("总 delta 不应为正，got %d", sum)
	}
}

// 爆冷（低分选手赢高分选手）应大涨；被爆的高分应掉。
func TestCalculate_Upset(t *testing.T) {
	res := Calculate([]Player{
		{Rating: 1200, Rank: 1}, // 弱者夺冠
		{Rating: 2200, Rank: 2}, // 强者屈居第二
	})
	if res[0].Delta <= 0 {
		t.Errorf("爆冷者应涨，got %d", res[0].Delta)
	}
	if res[1].Delta >= 0 {
		t.Errorf("被爆者应掉，got %d", res[1].Delta)
	}
	if res[0].Delta < -res[1].Delta {
		// 弱者夺冠的涨幅应可观（不强求严格大小，仅健全性）
		t.Logf("upset delta: winner +%d, loser %d", res[0].Delta, res[1].Delta)
	}
}

// NewRating = Rating + Delta 恒成立。
func TestCalculate_NewRatingConsistent(t *testing.T) {
	players := []Player{
		{Rating: 1500, Rank: 1}, {Rating: 1500, Rank: 2}, {Rating: 1500, Rank: 3},
	}
	for i, r := range Calculate(players) {
		if r.NewRating != players[i].Rating+r.Delta {
			t.Errorf("NewRating 不一致：%d != %d+%d", r.NewRating, players[i].Rating, r.Delta)
		}
	}
}
