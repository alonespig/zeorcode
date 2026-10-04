package judge

// SubmissionEventInfo 是判题进度事件中的提交快照。
// 实时推送端点只传递状态，不携带源代码，避免绕过提交详情接口的权限控制。
type SubmissionEventInfo struct {
	ID         int64  `json:"id"`
	Language   string `json:"language"`
	Status     int    `json:"status"`
	TimeUsed   int64  `json:"time"`
	MemoryUsed int64  `json:"memory"`
	CreatedAt  string `json:"createdAt"`
}

// SubmissionCaseResult 单个测试点的评测结果。
type SubmissionCaseResult struct {
	ID         int   `json:"id"`
	Status     int   `json:"status"`
	TimeUsed   int64 `json:"time"`
	MemoryUsed int64 `json:"memory"`
}
