package service

import (
	"context"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"
)

type NotificationService struct {
	repo     *repository.NotificationRepo
	userRepo *repository.UserRepo
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
func (s *NotificationService) Broadcast(ctx context.Context, req *dto.BroadcastReq) error {
	ids, err := s.userRepo.AllUserIDs(ctx)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	ns := make([]*model.Notification, 0, len(ids))
	for _, id := range ids {
		ns = append(ns, &model.Notification{
			RecipientID: id,
			Type:        "system",
			Title:       req.Title,
			Content:     req.Content,
			Link:        req.Link,
			SourceType:  "system",
		})
	}
	if err := s.repo.CreateInBatches(ctx, ns); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *NotificationService) List(ctx context.Context, recipientID int64, typ string, page, pageSize int) (*dto.NotificationListResp, error) {
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
	resp := &dto.NotificationListResp{Total: total, List: make([]dto.NotificationItem, 0, len(list))}
	for _, n := range list {
		item := dto.NotificationItem{
			ID:         n.ID,
			Type:       n.Type,
			Title:      n.Title,
			Content:    n.Content,
			Link:       n.Link,
			SourceType: n.SourceType,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if u, ok := userMap[n.ActorID]; ok {
			item.Actor = &dto.NotifActor{ID: u.UID, Username: u.Username, Avatar: avatarOr(u.Avatar), Rating: u.Rating}
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, recipientID int64) (*dto.UnreadCountResp, error) {
	m, err := s.repo.UnreadCountByType(ctx, recipientID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.UnreadCountResp{
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
