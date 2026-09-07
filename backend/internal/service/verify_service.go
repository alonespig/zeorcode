package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/mail"
	"zoj/internal/repository"

	goredis "github.com/redis/go-redis/v9"
)

// 验证码场景
const (
	SceneRegister = "register" // 注册验证：邮箱必须未被占用
	SceneReset    = "reset"    // 找回密码：邮箱必须已注册
	SceneBind     = "bind"     // 换绑/绑定邮箱：新邮箱必须未被占用
)

const (
	codeTTL       = 5 * time.Minute  // 验证码有效期
	codeCooldown  = 60 * time.Second // 同邮箱+场景两次发码最小间隔
	verifyFailTTL = 10 * time.Minute
	verifyFailMax = 5
)

// VerifyService 邮箱验证码：生成、存 Redis、发信、校验。
type VerifyService struct {
	cache    *cache.Cache
	mailer   *mail.Mailer
	userRepo *repository.UserRepo
}

func NewVerifyService(cache *cache.Cache, mailer *mail.Mailer, userRepo *repository.UserRepo) *VerifyService {
	return &VerifyService{cache: cache, mailer: mailer, userRepo: userRepo}
}

func codeKey(scene, email string) string       { return fmt.Sprintf("verifycode:%s:%s", scene, email) }
func cooldownKey(scene, email string) string   { return fmt.Sprintf("verifycd:%s:%s", scene, email) }
func verifyFailKey(scene, email string) string { return fmt.Sprintf("verifyfail:%s:%s", scene, email) }
func normalizeEmail(email string) string       { return strings.ToLower(strings.TrimSpace(email)) }

func (s *VerifyService) ensureVerifyNotLocked(ctx context.Context, scene, email string) error {
	raw, err := s.cache.GetString(ctx, verifyFailKey(scene, email))
	if errors.Is(err, goredis.Nil) {
		return nil
	}
	if err != nil {
		return errcode.ErrRedis.Wrap(err)
	}
	attempts, err := strconv.Atoi(raw)
	if err == nil && attempts >= verifyFailMax {
		return errcode.ErrTooManyRequests.WithMsg("验证码错误次数过多，请10分钟后重试")
	}
	return nil
}

// SendCode 发验证码：冷却拦截 → 场景预检 → 生成存库 → 发信。
// 任一步失败都会把冷却/验证码清掉，让用户能立刻重试。
func (s *VerifyService) SendCode(ctx context.Context, scene, email string) error {
	email = normalizeEmail(email)
	switch scene {
	case SceneRegister, SceneReset, SceneBind:
	default:
		return errcode.ErrInvalidParams.WithMsg("非法验证码场景")
	}
	if err := s.ensureVerifyNotLocked(ctx, scene, email); err != nil {
		return err
	}

	// 1) 冷却：60 秒内同邮箱+场景只能发一次（先拦，避免刷 DB / 刷信）
	ok, err := s.cache.SetNX(ctx, cooldownKey(scene, email), 1, codeCooldown)
	if err == nil && !ok {
		return errcode.ErrTooManyRequests.WithMsg("验证码发送过于频繁，请稍后再试")
	}

	// 2) 场景预检：注册/换绑要求邮箱未占用；找回密码要求邮箱已注册
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		_ = s.cache.Delete(ctx, cooldownKey(scene, email))
		return errcode.ErrDatabase.Wrap(err)
	}
	switch scene {
	case SceneRegister:
		if exists {
			_ = s.cache.Delete(ctx, cooldownKey(scene, email))
			return errcode.ErrEmailExists
		}
	case SceneBind:
		if exists {
			_ = s.cache.Delete(ctx, cooldownKey(scene, email))
			return errcode.ErrEmailExists.WithMsg("该邮箱已被占用")
		}
	case SceneReset:
		if !exists {
			_ = s.cache.Delete(ctx, cooldownKey(scene, email))
			return errcode.ErrUserNotFound.WithMsg("该邮箱未注册")
		}
	}

	// 3) 生成 + 存库
	code := genCode()
	if err := s.cache.Set(ctx, codeKey(scene, email), code, codeTTL); err != nil {
		_ = s.cache.Delete(ctx, cooldownKey(scene, email))
		return errcode.ErrRedis.Wrap(err)
	}

	// 4) 发信
	subject, body := buildMail(scene, code)
	if err := s.mailer.Send(email, subject, body); err != nil {
		_ = s.cache.Delete(ctx, cooldownKey(scene, email), codeKey(scene, email))
		return errcode.ErrInternal.WithMsg("邮件发送失败，请稍后再试或检查邮箱是否正确").Wrap(err)
	}
	return nil
}

// Verify 校验验证码，成功后立即删除（一次性）。
func (s *VerifyService) Verify(ctx context.Context, scene, email, code string) error {
	email = normalizeEmail(email)
	if code == "" {
		return errcode.ErrInvalidParams.WithMsg("请输入验证码")
	}
	if err := s.ensureVerifyNotLocked(ctx, scene, email); err != nil {
		return err
	}
	saved, err := s.cache.GetString(ctx, codeKey(scene, email))
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return errcode.ErrVerificationCodeExpired
		}
		return errcode.ErrRedis.Wrap(err)
	}
	if saved == "" {
		return errcode.ErrVerificationCodeExpired
	}
	if saved != code {
		attempts, incrErr := s.cache.RateLimitIncr(ctx, verifyFailKey(scene, email), verifyFailTTL)
		if incrErr != nil {
			return errcode.ErrRedis.Wrap(incrErr)
		}
		if attempts >= verifyFailMax {
			_ = s.cache.Delete(ctx, codeKey(scene, email))
			return errcode.ErrTooManyRequests.WithMsg("验证码错误次数过多，请10分钟后重试")
		}
		return errcode.ErrVerificationCodeInvalid
	}
	_ = s.cache.Delete(ctx, codeKey(scene, email), verifyFailKey(scene, email))
	return nil
}

// genCode 生成 6 位数字验证码（crypto/rand）。
func genCode() string {
	const digits = "0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// rand 基本不会失败；兜底用时间尾数拼一个，避免 panic
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	for i := range b {
		b[i] = digits[int(b[i])%10]
	}
	return string(b)
}

// buildMail 按场景生成邮件标题与 HTML 正文。
func buildMail(scene, code string) (subject, body string) {
	action := map[string]string{
		SceneRegister: "注册 ZOJ 账号",
		SceneReset:    "重置 ZOJ 登录密码",
		SceneBind:     "绑定新邮箱",
	}[scene]

	subject = "【ZOJ】验证码 " + code
	body = fmt.Sprintf(`<div style="max-width:480px;margin:0 auto;font-family:-apple-system,'Segoe UI',sans-serif;color:#1f2430;">
  <h2 style="font-size:18px;font-weight:600;margin:0 0 16px;">ZOJ 邮箱验证</h2>
  <p style="font-size:14px;color:#5b6470;margin:0 0 12px;">你正在%s，本次验证码为：</p>
  <div style="font-size:30px;font-weight:700;letter-spacing:8px;color:#20b2aa;margin:8px 0 16px;">%s</div>
  <p style="font-size:13px;color:#9aa1ab;margin:0;">验证码 5 分钟内有效，请勿泄露给他人。如非本人操作，请忽略此邮件。</p>
</div>`, action, code)
	return subject, body
}
