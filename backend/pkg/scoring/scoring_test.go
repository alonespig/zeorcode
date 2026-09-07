package scoring

import "testing"

func TestCFScore(t *testing.T) {
	cases := []struct {
		name                string
		x, tt, d, w         int
		accepted            bool
		want                int
	}{
		{"未AC返回0", 500, 30, 120, 0, false, 0},
		{"t=0满分", 500, 0, 120, 0, true, 500},
		{"赛末0.52x", 500, 120, 120, 0, true, 260},  // 500-floor(0.48*500)=500-240=260
		{"半程一次错交", 500, 60, 120, 1, true, 330}, // 500-120-50=330
		{"多次错交触发30%下限", 500, 100, 120, 6, true, 150}, // raw 很低 → 0.3*500=150
		{"错交也不低于30%", 500, 0, 120, 20, true, 150},     // 500-1000 → 下限 150
		{"大分值t=0", 3000, 0, 150, 0, true, 3000},
		{"x=0返回0", 0, 0, 120, 0, true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CFScore(c.x, c.tt, c.d, c.w, c.accepted)
			if got != c.want {
				t.Errorf("CFScore(x=%d,t=%d,d=%d,w=%d,ac=%v)=%d, want %d",
					c.x, c.tt, c.d, c.w, c.accepted, got, c.want)
			}
		})
	}
}

// 越晚提交得分不增（单调不升）。
func TestCFScore_MonotoneInTime(t *testing.T) {
	prev := CFScore(1000, 0, 120, 0, true)
	for tt := 1; tt <= 120; tt++ {
		cur := CFScore(1000, tt, 120, 0, true)
		if cur > prev {
			t.Fatalf("t=%d 得分 %d 反而高于前一分钟 %d", tt, cur, prev)
		}
		prev = cur
	}
}

// 任意 AC 情况下不低于 30%。
func TestCFScore_FloorAlways(t *testing.T) {
	x := 1500
	floor := 450
	for tt := 0; tt <= 120; tt++ {
		for w := 0; w <= 10; w++ {
			if got := CFScore(x, tt, 120, w, true); got < floor {
				t.Fatalf("t=%d w=%d 得分 %d 低于下限 %d", tt, w, got, floor)
			}
		}
	}
}
