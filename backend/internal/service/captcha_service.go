package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"

	"github.com/mojocn/base64Captcha"
	goredis "github.com/redis/go-redis/v9"
)

const captchaTTL = 5 * time.Minute

var loginCaptchaDriver = base64Captcha.NewDriverDigit(44, 132, 5, 0.7, 50)

type CaptchaStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	GetAndDelete(ctx context.Context, key string) (string, error)
}

type CaptchaService struct {
	store CaptchaStore
}

func NewCaptchaService(store *cache.Cache) *CaptchaService {
	return &CaptchaService{store: store}
}

func captchaKey(id string) string {
	return "captcha:login:" + id
}

func newCaptchaID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// Generate 生成登录验证码并将答案保存到 Redis，不使用依赖库的进程内全局 Store。
func (s *CaptchaService) Generate(ctx context.Context) (*dto.CaptchaResp, error) {
	id, err := newCaptchaID()
	if err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}
	_, content, answer := loginCaptchaDriver.GenerateSpecificIdQuestionAnswer(id)
	item, err := loginCaptchaDriver.DrawCaptcha(content)
	if err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}
	if err := s.store.Set(ctx, captchaKey(id), answer, captchaTTL); err != nil {
		return nil, errcode.ErrRedis.Wrap(err)
	}
	return &dto.CaptchaResp{
		ID:    id,
		Image: item.EncodeB64string(),
	}, nil
}

// Verify 原子消费验证码。无论答案是否正确，同一个验证码都不能再次使用。
func (s *CaptchaService) Verify(ctx context.Context, id, answer string) error {
	value, err := s.store.GetAndDelete(ctx, captchaKey(strings.TrimSpace(id)))
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return errcode.ErrCaptchaInvalid
		}
		return errcode.ErrRedis.Wrap(err)
	}
	if !strings.EqualFold(value, strings.TrimSpace(answer)) {
		return errcode.ErrCaptchaInvalid
	}
	return nil
}
