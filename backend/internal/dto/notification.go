package dto

// NotifActor 通知触发者（供前端按段位染色）
type NotifActor struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Rating   int    `json:"rating"`
}

type NotificationItem struct {
	ID        int64       `json:"id"`
	Type      string      `json:"type"`
	Title     string      `json:"title,omitempty"`
	Content   string      `json:"content,omitempty"`
	Link       string      `json:"link"`
	SourceType string      `json:"sourceType,omitempty"` // post / comment / contest …（前端区分点赞帖子/评论）
	IsRead     bool        `json:"isRead"`
	CreatedAt string      `json:"createdAt"`
	Actor     *NotifActor `json:"actor,omitempty"` // 系统/rating 无触发者
}

type NotificationListResp struct {
	Total int64              `json:"total"`
	List  []NotificationItem `json:"list"`
}

// UnreadCountResp 铃铛/各 tab 角标
type UnreadCountResp struct {
	Comment int64 `json:"comment"`
	Reply   int64 `json:"reply"`
	Like    int64 `json:"like"`
	System  int64 `json:"system"`
	Rating  int64 `json:"rating"`
	Total   int64 `json:"total"`
}

// MarkReadReq 标记已读：ID>0 单条；Type 非空该类型全部；都空=全部
type MarkReadReq struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// BroadcastReq 管理员广播系统通知
type BroadcastReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Link    string `json:"link"`
}
