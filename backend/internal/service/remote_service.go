package service

import (
	"context"
	"errors"

	"zoj/internal/common/errcode"
	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/remoteoj"
	"zoj/pkg/util"

	"gorm.io/gorm"
)

type RemoteService struct {
	accRepo *repository.RemoteAccountRepo
}

func NewRemoteService(accRepo *repository.RemoteAccountRepo) *RemoteService {
	return &RemoteService{accRepo: accRepo}
}

// ListOJ 返回支持的远程 OJ 列表
func (s *RemoteService) ListOJ(ctx context.Context) []string {
	return remoteoj.Supported()
}

// CrawlProblem 拉取指定 OJ 的题目（自动带上该 OJ 的启用账号 cookie）
func (s *RemoteService) CrawlProblem(ctx context.Context, ojName, pid string) (*dto.RemoteProblemResp, error) {
	oj, ok := remoteoj.Get(ojName)
	if !ok {
		return nil, errcode.ErrRemoteOJUnsupported
	}

	p, err := oj.CrawlProblem(s.loadAccount(ctx, ojName), pid)
	if err != nil {
		return nil, errcode.ErrRemoteProblemUnavailable.Wrap(err)
	}

	samples := make([]dto.ProblemSample, 0, len(p.Samples))
	for _, sp := range p.Samples {
		samples = append(samples, dto.ProblemSample{
			Input:   sp.Input,
			Output:  sp.Output,
			Explain: sp.Explain,
		})
	}

	return &dto.RemoteProblemResp{
		OJ:              p.OJ,
		RemoteProblemID: p.RemoteProblemID,
		Title:           p.Title,
		TimeLimit:       p.TimeLimit,
		MemoryLimit:     p.MemoryLimit,
		Description:     p.Description,
		InputFormat:     p.InputFormat,
		OutputFormat:    p.OutputFormat,
		Samples:         samples,
		Hint:            p.Hint,
		Source:          p.Source,
	}, nil
}

// loadAccount 取该 OJ 的启用账号并解密 cookie；取不到/解密失败则返回 nil(走匿名/配置回退)。
func (s *RemoteService) loadAccount(ctx context.Context, ojName string) *remoteoj.RemoteAccount {
	m, err := s.accRepo.FirstEnabledByOJ(ctx, ojName)
	if err != nil {
		return nil // 无启用账号，按匿名处理
	}
	secret, err := util.Decrypt(m.Secret)
	if err != nil {
		logger.Warnw("decrypt remote account secret failed", "oj", ojName, "accountID", m.ID, "err", err)
		return nil
	}
	acc := &remoteoj.RemoteAccount{ID: m.ID, OJ: m.OJ, Username: m.Username, Valid: m.Valid}
	if m.AuthType == "cookie" {
		acc.Cookie = secret
	} else {
		acc.Password = secret
	}
	return acc
}

// ==================== 远程账号管理 ====================

func (s *RemoteService) ListAccounts(ctx context.Context) ([]dto.RemoteAccountItem, error) {
	list, err := s.accRepo.List(ctx)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	items := make([]dto.RemoteAccountItem, 0, len(list))
	for _, m := range list {
		items = append(items, dto.RemoteAccountItem{
			ID:        m.ID,
			OJ:        m.OJ,
			AuthType:  m.AuthType,
			Username:  m.Username,
			HasSecret: m.Secret != "",
			Enabled:   m.Enabled,
			Valid:     m.Valid,
			Busy:      m.Busy,
			CreatedAt: m.CreatedAt,
		})
	}
	return items, nil
}

func (s *RemoteService) CreateAccount(ctx context.Context, req *dto.CreateRemoteAccountReq) (int64, error) {
	enc, err := encryptSecret(req.Secret)
	if err != nil {
		return 0, err
	}
	id, err := s.accRepo.Create(ctx, &model.RemoteAccount{
		OJ:       req.OJ,
		AuthType: req.AuthType,
		Username: req.Username,
		Secret:   enc,
		Enabled:  req.Enabled,
	})
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return id, nil
}

func (s *RemoteService) UpdateAccount(ctx context.Context, id int64, req *dto.UpdateRemoteAccountReq) error {
	fields := map[string]any{
		"auth_type": req.AuthType,
		"username":  req.Username,
		"enabled":   req.Enabled,
	}
	// 留空表示不改凭证；改了则重新加密并重置有效性
	if req.Secret != "" {
		enc, err := encryptSecret(req.Secret)
		if err != nil {
			return err
		}
		fields["secret"] = enc
		fields["valid"] = false
	}
	if err := s.accRepo.Update(ctx, id, fields); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRemoteAccountNotFound
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *RemoteService) DeleteAccount(ctx context.Context, id int64) error {
	if err := s.accRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRemoteAccountNotFound
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func encryptSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	enc, err := util.Encrypt(plain)
	if err != nil {
		return "", errcode.ErrInternal.WithMsg("凭证加密失败").Wrap(err)
	}
	return enc, nil
}
