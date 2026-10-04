package handler

import (
	"strconv"

	"zoj/internal/http/dto"
	"zoj/internal/service"
	"zoj/pkg/errcode"

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
	resp, err := n.srv.List(c.Request.Context(), uid, c.Query("type"), page, pageSize)
	if err != nil {
		return nil, err
	}
	return toNotificationListResp(resp), nil
}

// UnreadCount GET /api/notifications/unread-count
func (n *NotificationController) UnreadCount(c *gin.Context) (any, error) {
	uid := notifUID(c)
	if uid == 0 {
		return nil, errcode.ErrUnauthorized
	}
	resp, err := n.srv.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		return nil, err
	}
	return toUnreadCountResp(resp), nil
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
	return nil, n.srv.Broadcast(c.Request.Context(), service.BroadcastParams{
		Title: req.Title, Content: req.Content, Link: req.Link,
	})
}

func toNotificationListResp(resp *service.NotificationList) *dto.NotificationListResp {
	items := make([]dto.NotificationItem, 0, len(resp.List))
	for _, it := range resp.List {
		item := dto.NotificationItem{
			ID:         it.ID,
			Type:       it.Type,
			Title:      it.Title,
			Content:    it.Content,
			Link:       it.Link,
			SourceType: it.SourceType,
			IsRead:     it.IsRead,
			CreatedAt:  it.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if it.Actor != nil {
			item.Actor = &dto.NotifActor{
				ID:       it.Actor.ID,
				Username: it.Actor.Username,
				Avatar:   it.Actor.Avatar,
				Rating:   it.Actor.Rating,
			}
		}
		items = append(items, item)
	}
	return &dto.NotificationListResp{Total: resp.Total, List: items}
}

func toUnreadCountResp(u *service.UnreadCount) *dto.UnreadCountResp {
	return &dto.UnreadCountResp{
		Comment: u.Comment,
		Reply:   u.Reply,
		Like:    u.Like,
		System:  u.System,
		Rating:  u.Rating,
		Total:   u.Total,
	}
}
