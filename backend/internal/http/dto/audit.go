package dto

type AuditLogListReq struct {
	PageForm
	Keyword string `form:"q" binding:"max=100"`
	Method  string `form:"method" binding:"omitempty,oneof=POST PUT PATCH DELETE"`
	Success *bool  `form:"success"`
}

type AuditLogItemResp struct {
	ID        int64  `json:"id"`
	ActorID   int64  `json:"actorID"`
	ActorName string `json:"actorName"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Target    string `json:"target"`
	ClientIP  string `json:"clientIP"`
	Success   bool   `json:"success"`
	Code      int    `json:"code"`
	CreatedAt string `json:"createdAt"`
}

type AuditLogListResp struct {
	Total int64              `json:"total"`
	List  []AuditLogItemResp `json:"list"`
}
