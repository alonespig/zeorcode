package service

import (
	"context"
	"time"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
)

type NotificationService struct {
	repo     *repository.NotificationRepo
	userRepo *repository.UserRepo
}

// NotifActor 通知触发者。
type NotifActor struct {
	ID       int64
	Username string
	Avatar   string
	Rating   int
}

// NotificationItem 一条通知及其触发者信息。
type NotificationItem struct {
	ID         int64
	Type       string
	Title      string
	Content    string
	Link       string
	SourceType string
	IsRead     bool
	CreatedAt  time.Time
	Actor      *NotifActor
}

// NotificationList 通知列表结果。
type NotificationList struct {
	Total int64
	List  []NotificationItem
}

// UnreadCount 未读通知计数。
type UnreadCount struct {
	Comment int64
	Reply   int64
	Like    int64
	System  int64
	Rating  int64
	Total   int64
}

// BroadcastParams 广播系统通知的输入。
type BroadcastParams struct {
	ActorID int64
	Title   string
	Content string
	Link    string
}

type BroadcastHistoryItem struct {
	ID             int64
	ActorID        int64
	ActorName      string
	Title          string
	Content        string
	Link           string
	RecipientCount int
	CreatedAt      time.Time
}

type BroadcastHistory struct {
	Total int64
	List  []BroadcastHistoryItem
}

func NewNotificationService(repo *repository.NotificationRepo, userRepo *repository.UserRepo) *NotificationService {
	return &NotificationService{repo: repo, userRepo: userRepo}
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// Notify 发一条通知（副作用：不给自己发、落库失败不阻断主流程）。
func (s *NotificationService) Notify(ctx context.Context, recipientID, actorID int64, typ, title, content, link, srcType string, srcID int64) {
	if recipientID <= 0 || recipientID == actorID {
		return
	}
	_ = s.repo.Create(ctx, &model.Notification{
		RecipientID: recipientID,
		Type:        typ,
		ActorID:     actorID,
		Title:       title,
		Content:     truncateRunes(content, 80),
		Link:        link,
		SourceType:  srcType,
		SourceID:    srcID,
	})
}

// Broadcast 管理员系统通知：扇出给全体用户。
func (s *NotificationService) Broadcast(ctx context.Context, params BroadcastParams) error {
	ids, err := s.userRepo.AllUserIDs(ctx)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	ns := make([]*model.Notification, 0, len(ids))
	for _, id := range ids {
		ns = append(ns, &model.Notification{
			RecipientID: id,
			Type:        "system",
			Title:       params.Title,
			Content:     params.Content,
			Link:        params.Link,
			SourceType:  "system",
		})
	}
	broadcast := &model.SystemBroadcast{
		ActorID: params.ActorID, Title: params.Title, Content: params.Content,
		Link: params.Link, RecipientCount: len(ids),
	}
	if err := s.repo.CreateBroadcast(ctx, broadcast, ns); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *NotificationService) BroadcastHistory(ctx context.Context, page, pageSize int) (*BroadcastHistory, error) {
	rows, total, err := s.repo.ListBroadcasts(ctx, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result := &BroadcastHistory{Total: total, List: make([]BroadcastHistoryItem, 0, len(rows))}
	for _, row := range rows {
		result.List = append(result.List, BroadcastHistoryItem{
			ID: row.ID, ActorID: row.ActorID, ActorName: row.ActorName,
			Title: row.Title, Content: row.Content, Link: row.Link,
			RecipientCount: row.RecipientCount, CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}

func (s *NotificationService) List(ctx context.Context, recipientID int64, typ string, page, pageSize int) (*NotificationList, error) {
	list, total, err := s.repo.List(ctx, recipientID, typ, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	actorIDs := make([]int64, 0, len(list))
	for _, n := range list {
		if n.ActorID > 0 {
			actorIDs = append(actorIDs, n.ActorID)
		}
	}
	userMap := map[int64]model.User{}
	if len(actorIDs) > 0 {
		if users, e := s.userRepo.FindByIDs(ctx, actorIDs); e == nil {
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}
	resp := &NotificationList{Total: total, List: make([]NotificationItem, 0, len(list))}
	for _, n := range list {
		item := NotificationItem{
			ID:         n.ID,
			Type:       n.Type,
			Title:      n.Title,
			Content:    n.Content,
			Link:       n.Link,
			SourceType: n.SourceType,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt,
		}
		if u, ok := userMap[n.ActorID]; ok {
			item.Actor = &NotifActor{ID: u.UID, Username: u.Username, Avatar: avatarOr(u.Avatar), Rating: u.Rating}
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, recipientID int64) (*UnreadCount, error) {
	m, err := s.repo.UnreadCountByType(ctx, recipientID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &UnreadCount{
		Comment: m["comment"],
		Reply:   m["reply"],
		Like:    m["like"],
		System:  m["system"],
		Rating:  m["rating"],
	}
	resp.Total = resp.Comment + resp.Reply + resp.Like + resp.System + resp.Rating
	return resp, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, recipientID, id int64, typ string) error {
	if err := s.repo.MarkRead(ctx, recipientID, id, typ); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}
