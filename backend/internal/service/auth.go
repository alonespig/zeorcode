package service

import (
	"context"
	"errors"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/internal/infra/session"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Claims 认证后的用户身份信息。
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// userIdentityReader 提供当前用户身份，用于每次请求校验实时角色与封禁状态。
type userIdentityReader interface {
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
}

// TokenService 鉴权组件：JWT 签发/校验 + 登录态存储（session）。
// 通过 DI 注入 *session.Session 与用户仓储。
type TokenService struct {
	session *session.Session
	users   userIdentityReader
	secret  []byte
	expire  time.Duration
}

func NewTokenService(s *session.Session, users userIdentityReader) *TokenService {
	return &TokenService{
		session: s,
		users:   users,
		secret:  []byte(viper.GetString("jwt.secret")),
		expire:  time.Duration(viper.GetInt("jwt.expire")) * time.Hour,
	}
}

// ExpireDuration 返回 JWT 与登录 Cookie 共用的有效期。
func (t *TokenService) ExpireDuration() time.Duration {
	return t.expire
}

// GenerateToken 生成带过期时间的 Token，并写入 session。
func (t *TokenService) GenerateToken(userID int64, username string, role int) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bluebell",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	if err := t.session.Save(ctx, userID, token, t.expire); err != nil {
		return "", err
	}
	return token, nil
}

// VerifyToken 解析并校验 Token：区分过期/非法，并二次确认登录态仍在 session 中。
func (t *TokenService) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errcode.ErrTokenExpired
		}
		return nil, errcode.ErrInvalidToken.Wrap(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errcode.ErrInvalidToken
	}

	stored, err := t.session.Get(context.Background(), claims.UserID)
	if err != nil {
		logger.Errorw("failed to get token from session", "userID", claims.UserID, "err", err)
		return nil, errcode.ErrTokenExpired
	}
	if stored != tokenString {
		return nil, errcode.ErrTokenExpired
	}
	return claims, nil
}

// CurrentRole 返回数据库里的实时角色。
// JWT 中的 role 只作为签发时的快照，权限判断一律以这里为准，
// 这样降权或封禁即使撤销会话失败，旧 token 也会立即失去管理员身份。
func (t *TokenService) CurrentRole(ctx context.Context, userID int64) (int, error) {
	user, err := t.users.GetUserByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrUserNotFound
	}
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if user.Status == model.UserStatusBanned {
		return 0, errcode.ErrUserDisabled
	}
	return user.Role, nil
}

// Revoke 使某用户的登录态失效（登出）。
func (t *TokenService) Revoke(ctx context.Context, userID int64) error {
	return t.session.Delete(ctx, userID)
}
