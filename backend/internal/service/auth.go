package service

import (
	"context"
	"errors"
	"time"

	"zoj/internal/infra/logger"
	"zoj/internal/infra/session"
	"zoj/pkg/errcode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// Claims 认证后的用户身份信息。
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// TokenService 鉴权组件：JWT 签发/校验 + 登录态存储（session）。
// 通过 DI 注入 *session.Session，不再依赖全局 Redis 客户端。
type TokenService struct {
	session *session.Session
	secret  []byte
	expire  time.Duration
}

func NewTokenService(s *session.Session) *TokenService {
	return &TokenService{
		session: s,
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
		return nil, errcode.ErrTokenExpired.WithMsg("登录状态已失效")
	}
	if stored != tokenString {
		return nil, errcode.ErrTokenExpired.WithMsg("登录状态已失效")
	}
	return claims, nil
}

// Revoke 使某用户的登录态失效（登出）。
func (t *TokenService) Revoke(ctx context.Context, userID int64) error {
	return t.session.Delete(ctx, userID)
}
