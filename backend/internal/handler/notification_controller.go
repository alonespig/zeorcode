package handler

import (
	"strconv"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	srv *service.NotificationService
}

func NewNotificationController(srv *service.NotificationService) *NotificationController {
	return &NotificationController{srv: srv}
}

func notifUID(c *gin.Context) int64 {
	v, _ := c.Get("userID")
	id, _ := v.(int64)
	return id
}

// List GET /api/notifications?type=&page=&pageSize=
func (n *NotificationController) List(c *gin.Context) (any, error) {
	uid := notifUID(c)
	if uid == 0 {
		return nil, errcode.ErrUnauthorized
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	return n.srv.List(c.Request.Context(), uid, c.Query("type"), page, pageSize)
}

// UnreadCount GET /api/notifications/unread-count
func (n *NotificationController) UnreadCount(c *gin.Context) (any, error) {
	uid := notifUID(c)
	if uid == 0 {
		return nil, errcode.ErrUnauthorized
	}
	return n.srv.UnreadCount(c.Request.Context(), uid)
}

// MarkRead POST /api/notifications/read  body: {id} 单条 / {type} 该类型 / 空 全部
func (n *NotificationController) MarkRead(c *gin.Context) (any, error) {
	uid := notifUID(c)
	if uid == 0 {
		return nil, errcode.ErrUnauthorized
	}
	var req dto.MarkReadReq
	_ = c.ShouldBindJSON(&req) // 允许空 body（=全部已读）
	return nil, n.srv.MarkRead(c.Request.Context(), uid, req.ID, req.Type)
}

// Broadcast POST /api/admin/notifications  管理员广播系统通知
func (n *NotificationController) Broadcast(c *gin.Context) (any, error) {
	var req dto.BroadcastReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return nil, n.srv.Broadcast(c.Request.Context(), &req)
}
