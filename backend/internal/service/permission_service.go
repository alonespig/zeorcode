package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/infra/cache"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"
)

// permissionCacheTTL 是权限集合的兜底缓存时间；权限变更时会通过版本号切换缓存键。
const permissionCacheTTL = time.Minute

func permissionCacheGenKey(userID int64) string {
	return fmt.Sprintf("authz:gen:%d", userID)
}

func permissionCacheKey(userID int64, generation string) string {
	return fmt.Sprintf("authz:user:%d:%s", userID, generation)
}

type permissionStore interface {
	List(ctx context.Context, userID int64) ([]string, error)
	Replace(ctx context.Context, userID int64, permissions []string) error
	DeleteByUser(ctx context.Context, userID int64) error
}

type permissionCache interface {
	GetString(ctx context.Context, key string) (string, error)
	GetJSON(ctx context.Context, key string, dest any) (bool, error)
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error
	Incr(ctx context.Context, key string) (int64, error)
}

// PermissionService 负责管理员权限的读取、校验与分配。
// 超级管理员拥有目录中的全部权限，无需查询授权表。
type PermissionService struct {
	repo  permissionStore
	users *repository.UserRepo
	cache permissionCache
}

func NewPermissionService(
	repo *repository.AdminPermissionRepo,
	users *repository.UserRepo,
	c *cache.Cache,
) *PermissionService {
	return &PermissionService{repo: repo, users: users, cache: c}
}

// List 返回某个内部用户 ID 被授予的权限集合。
func (s *PermissionService) List(ctx context.Context, userID int64) ([]string, error) {
	// 只使用版本明确的缓存。版本缺失或读取失败时直接回源，避免命中旧的无版本缓存。
	generation, genErr := s.cache.GetString(ctx, permissionCacheGenKey(userID))
	var key string
	if genErr == nil && generation != "" {
		key = permissionCacheKey(userID, generation)
		var cached []string
		if ok, err := s.cache.GetJSON(ctx, key, &cached); err == nil && ok {
			return activePermissions(cached), nil
		}
	}

	permissions, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	permissions = activePermissions(permissions)
	if key != "" {
		_ = s.cache.SetJSON(ctx, key, permissions, permissionCacheTTL)
	}
	return permissions, nil
}

// activePermissions 忽略已从目录移除的旧授权，避免旧记录进入鉴权或权限编辑表单。
func activePermissions(permissions []string) []string {
	return ensureDashboard(slices.DeleteFunc(permissions, func(permission string) bool {
		return !authz.IsValid(permission)
	}))
}

// ensureDashboard 保证管理员至少拥有工作台权限：
// 没有这条权限时连后台首页都进不去，默认给上是更合理的兜底。
func ensureDashboard(permissions []string) []string {
	if slices.Contains(permissions, authz.DashboardRead) {
		return permissions
	}
	return append(permissions, authz.DashboardRead)
}

// PermissionsFor 根据角色返回有效权限集合：
// 超级管理员拥有全部权限，普通用户没有后台权限，admin 查授权表。
func (s *PermissionService) PermissionsFor(ctx context.Context, userID int64, role int) ([]string, error) {
	switch role {
	case model.RoleSuperAdmin:
		return authz.All(), nil
	case model.RoleAdmin:
		return s.List(ctx, userID)
	default:
		return []string{}, nil
	}
}

// HasAny 判断用户是否拥有任意一个给定权限。未指定权限时拒绝访问。
func (s *PermissionService) HasAny(ctx context.Context, userID int64, role int, permissions ...string) (bool, error) {
	if len(permissions) == 0 {
		return false, nil
	}
	if role == model.RoleSuperAdmin {
		return slices.ContainsFunc(permissions, authz.IsValid), nil
	}
	if role != model.RoleAdmin {
		return false, nil
	}
	granted, err := s.List(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, permission := range permissions {
		if slices.Contains(granted, permission) {
			return true, nil
		}
	}
	return false, nil
}

// GetByUID 返回对外用户号为 uid 的用户角色与已授权限，供权限配置界面回填。
func (s *PermissionService) GetByUID(ctx context.Context, uid int64) ([]string, int, error) {
	user, err := s.resolveUser(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	if user.Role != model.RoleAdmin {
		return []string{}, user.Role, nil
	}
	permissions, err := s.List(ctx, user.ID)
	return permissions, user.Role, err
}

// SetByUID 给对外用户号为 uid 的 admin 整组替换权限。
// 只允许给 admin 配置；super_admin 默认拥有全部权限，普通用户没有后台入口。
func (s *PermissionService) SetByUID(ctx context.Context, uid int64, permissions []string) error {
	user, err := s.resolveUser(ctx, uid)
	if err != nil {
		return err
	}
	if user.Role != model.RoleAdmin {
		return errcode.ErrOperationDenied.WithMsg("只能给管理员配置权限")
	}

	normalized := make([]string, 0, len(permissions))
	seen := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		if !authz.IsValid(permission) {
			return errcode.ErrInvalidParams.WithMsg("未知权限点：" + permission)
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		normalized = append(normalized, permission)
	}
	normalized = ensureDashboard(normalized)

	if err := s.repo.Replace(ctx, user.ID, normalized); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return s.Invalidate(ctx, user.ID)
}

// ClearForUser 收回某个用户的全部权限（角色降级时调用）。
func (s *PermissionService) ClearForUser(ctx context.Context, userID int64) error {
	if err := s.repo.DeleteByUser(ctx, userID); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return s.Invalidate(ctx, userID)
}

// Invalidate 通过递增版本号让旧缓存失效，下一次请求使用新键重新加载。
// 返回错误而不是静默忽略：递增失败时旧权限可能还会在 TTL 内生效，
// 调用方需要知道这次变更没有完全生效。
func (s *PermissionService) Invalidate(ctx context.Context, userID int64) error {
	if _, err := s.cache.Incr(ctx, permissionCacheGenKey(userID)); err != nil {
		return errcode.ErrRedis.Wrap(err)
	}
	return nil
}

func (s *PermissionService) resolveUser(ctx context.Context, uid int64) (*model.User, error) {
	id, err := s.users.ResolveIDByUID(ctx, uid)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}
	user, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}
	return user, nil
}
