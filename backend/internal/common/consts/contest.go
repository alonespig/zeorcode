package consts

// ContestType 比赛模式
type ContestType int

const (
	ContestACM ContestType = 1
	ContestOI  ContestType = 2
	ContestIOI ContestType = 3
	ContestCF  ContestType = 4 // Codeforces 动态分：分值随时间衰减 + 错交扣分 + 全过才得分
)

func (t ContestType) String() string {
	switch t {
	case ContestACM:
		return "ACM"
	case ContestOI:
		return "OI"
	case ContestIOI:
		return "IOI"
	case ContestCF:
		return "CF"
	default:
		return ""
	}
}

// ContestStatus 比赛状态
type ContestStatus int

const (
	ContestNotStarted ContestStatus = 0 // 未开始
	ContestRunning    ContestStatus = 1 // 进行中
	ContestFinished   ContestStatus = 2 // 已结束
)

func (s ContestStatus) String() string {
	switch s {
	case ContestNotStarted:
		return "未开始"
	case ContestRunning:
		return "进行中"
	case ContestFinished:
		return "已结束"
	default:
		return ""
	}
}
