package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/common/response"
	"zoj/internal/infra/session"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func getJWTSecret() []byte {
	return []byte(viper.GetString("jwt.secret"))
}

func getTokenExpireDuration() time.Duration {
	return time.Duration(viper.GetInt("jwt.expire")) * time.Hour
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// Auth 鉴权组件：JWT 签发/校验 + 登录态存储（session）。
// 通过 DI 注入 *session.Session，不再依赖全局 Redis 客户端。
type Auth struct {
	session *session.Session
}

func NewAuth(s *session.Session) *Auth {
	return &Auth{session: s}
}

// GenerateToken 生成带过期时间的 Token，并写入 session。
func (a *Auth) GenerateToken(userID int64, username string, role int) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(getTokenExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bluebell",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getJWTSecret())
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	if err := a.session.Save(ctx, userID, token, getTokenExpireDuration()); err != nil {
		return "", err
	}
	return token, nil
}

// Revoke 使某用户的登录态失效（登出）。
func (a *Auth) Revoke(ctx context.Context, userID int64) error {
	return a.session.Delete(ctx, userID)
}

// parseToken 解析并验证 Token，返回 *AppErr 区分过期/非法
func (a *Auth) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
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
	// 二次确认：检查该 token 是否仍在 session 中
	if !a.tokenValid(claims.UserID, tokenString) {
		return nil, errcode.ErrTokenExpired.WithMsg("登录状态已失效")
	}
	return claims, nil
}

func (a *Auth) tokenValid(userID int64, tokenString string) bool {
	stored, err := a.session.Get(context.Background(), userID)
	if err != nil {
		zap.S().Error("failed to get token from session", err)
		return false
	}
	return stored == tokenString
}

func (a *Auth) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := a.parseToken(tokenString)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (a *Auth) JWTAuthOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := a.parseToken(tokenString)
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("role", claims.Role)
		}

		c.Next()
	}
}

// AdminRequired 必须在 JWTAuth 之后挂，依赖 c.Get("role")；不涉及 session，保持无状态。
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		r, _ := role.(int)
		if r != 1 {
			response.Fail(c, errcode.ErrAdminRequired)
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractToken 从 Cookie 中提取 token
func extractToken(c *gin.Context) string {
	token, err := c.Cookie("token")
	if err != nil {
		return ""
	}
	return token
}

// SetTokenCookie 将 JWT 写入 HttpOnly Cookie
func SetTokenCookie(c *gin.Context, tokenString string) {
	expireSeconds := int(getTokenExpireDuration().Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", tokenString, expireSeconds, "/", "", false, true)
}

// ClearTokenCookie 清除 token cookie
func ClearTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", "", -1, "/", "", false, true)
}
