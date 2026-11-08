package service

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"
)

const permissionTestUserID = int64(42)

type fakePermissionStore struct {
	granted []string
	err     error
	reads   int
}

func (s *fakePermissionStore) Replace(_ context.Context, _ int64, permissions []string) error {
	s.granted = slices.Clone(permissions)
	return nil
}

func (s *fakePermissionStore) DeleteByUser(context.Context, int64) error {
	s.granted = nil
	return nil
}

func (s *fakePermissionStore) List(context.Context, int64) ([]string, error) {
	s.reads++
	return slices.Clone(s.granted), s.err
}

type fakePermissionCache struct {
	values        map[string]string
	generationErr error
	readErr       error
	writeErr      error
	incrErr       error
}

func (c *fakePermissionCache) GetString(_ context.Context, key string) (string, error) {
	return c.values[key], c.generationErr
}

func (c *fakePermissionCache) GetJSON(_ context.Context, key string, dest any) (bool, error) {
	if c.readErr != nil {
		return false, c.readErr
	}
	raw, ok := c.values[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal([]byte(raw), dest)
}

func (c *fakePermissionCache) SetJSON(_ context.Context, key string, value any, _ time.Duration) error {
	if c.writeErr != nil {
		return c.writeErr
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.values[key] = string(raw)
	return nil
}

func (c *fakePermissionCache) Incr(_ context.Context, key string) (int64, error) {
	if c.incrErr != nil {
		return 0, c.incrErr
	}
	var n int64
	if raw, ok := c.values[key]; ok {
		var err error
		n, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	n++
	c.values[key] = strconv.FormatInt(n, 10)
	return n, nil
}

func TestPermissionServiceHasAny(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name     string
		role     int
		required []string
		want     bool
	}{
		{name: "admin with no requirements", role: model.RoleAdmin},
		{name: "super admin with no requirements", role: model.RoleSuperAdmin},
		{name: "ordinary user", role: model.RoleNormal, required: []string{authz.ProblemRead}},
		{name: "super admin with a known permission", role: model.RoleSuperAdmin, required: []string{authz.ProblemWrite}, want: true},
		{name: "super admin with a removed permission", role: model.RoleSuperAdmin, required: []string{"problem.delete"}},
		{name: "any one grant is enough", role: model.RoleAdmin, required: []string{authz.ProblemWrite, authz.ProblemRead}, want: true},
		{name: "no matching grant", role: model.RoleAdmin, required: []string{authz.ProblemWrite}},
		{name: "default dashboard grant", role: model.RoleAdmin, required: []string{authz.DashboardRead}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PermissionService{}
			if tt.role == model.RoleAdmin && len(tt.required) != 0 {
				s.repo = &fakePermissionStore{granted: []string{authz.ProblemRead}}
				s.cache = &fakePermissionCache{}
			}
			got, err := s.HasAny(ctx, permissionTestUserID, tt.role, tt.required...)
			if err != nil || got != tt.want {
				t.Fatalf("HasAny() = %t, %v; want %t, nil", got, err, tt.want)
			}
		})
	}
}

func TestPermissionServiceCacheInvalidation(t *testing.T) {
	ctx := t.Context()
	repo := &fakePermissionStore{granted: []string{authz.ProblemWrite}}
	c := &fakePermissionCache{values: make(map[string]string)}
	s := &PermissionService{repo: repo, cache: c}

	if err := s.Invalidate(ctx, permissionTestUserID); err != nil {
		t.Fatal(err)
	}
	check := func(wantAllowed bool, wantReads int) {
		t.Helper()
		allowed, err := s.HasAny(ctx, permissionTestUserID, model.RoleAdmin, authz.ProblemWrite)
		if err != nil || allowed != wantAllowed || repo.reads != wantReads {
			t.Fatalf("HasAny() = %t, %v, with %d DB reads; want %t, nil, with %d reads", allowed, err, repo.reads, wantAllowed, wantReads)
		}
	}
	check(true, 1) // 首次读库并填充缓存。
	repo.granted = nil
	check(true, 1) // 当前版本仍读缓存。
	if err := s.ClearForUser(ctx, permissionTestUserID); err != nil {
		t.Fatal(err)
	}
	check(false, 2) // 切换版本后，即使旧缓存还在，也不再使用。
	check(false, 2)
}

func TestPermissionServiceCacheReadFallback(t *testing.T) {
	cacheErr := errors.New("redis unavailable")
	tests := []struct {
		name       string
		generation string
		genErr     error
		readErr    error
	}{
		{name: "version unavailable", genErr: cacheErr},
		{name: "version missing"},
		{name: "permission cache unavailable", generation: "5", readErr: cacheErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakePermissionStore{granted: []string{authz.ProblemRead}}
			c := &fakePermissionCache{
				values: map[string]string{
					permissionCacheGenKey(permissionTestUserID):             tt.generation,
					permissionCacheKey(permissionTestUserID, tt.generation): `["problem.write"]`,
				},
				generationErr: tt.genErr,
				readErr:       tt.readErr,
				writeErr:      cacheErr,
			}
			s := &PermissionService{repo: repo, cache: c}
			got, err := s.List(t.Context(), permissionTestUserID)
			if err != nil || repo.reads != 1 || !slices.Equal(got, []string{authz.ProblemRead, authz.DashboardRead}) {
				t.Fatalf("List() = %v, %v with %d DB reads; want fresh DB grants", got, err, repo.reads)
			}
		})
	}
}

func TestPermissionServiceIgnoresRemovedGrants(t *testing.T) {
	for _, fromCache := range []bool{false, true} {
		name := "database"
		if fromCache {
			name = "cache"
		}
		t.Run(name, func(t *testing.T) {
			c := &fakePermissionCache{values: map[string]string{
				permissionCacheGenKey(permissionTestUserID): "1",
			}}
			s := &PermissionService{cache: c}
			if fromCache {
				c.values[permissionCacheKey(permissionTestUserID, "1")] = `["problem.delete","contest.delete","problem.read"]`
			} else {
				s.repo = &fakePermissionStore{granted: []string{"problem.delete", "contest.delete", authz.ProblemRead}}
			}
			got, err := s.List(t.Context(), permissionTestUserID)
			if err != nil || !slices.Equal(got, []string{authz.ProblemRead, authz.DashboardRead}) {
				t.Fatalf("List() = %v, %v; want only active grants", got, err)
			}
		})
	}
}

func TestPermissionServiceBackendFailures(t *testing.T) {
	ctx := t.Context()
	dbErr := errors.New("database unavailable")
	cacheErr := errors.New("redis unavailable")
	s := &PermissionService{
		repo:  &fakePermissionStore{err: dbErr},
		cache: &fakePermissionCache{generationErr: cacheErr, incrErr: cacheErr},
	}
	allowed, err := s.HasAny(ctx, permissionTestUserID, model.RoleAdmin, authz.ProblemRead)
	var appErr *errcode.AppErr
	if allowed || !errors.Is(err, dbErr) || !errors.As(err, &appErr) || appErr.Code != errcode.DatabaseError {
		t.Fatalf("HasAny() = %t, %v; want deny with database error", allowed, err)
	}
	err = s.Invalidate(ctx, permissionTestUserID)
	if !errors.Is(err, cacheErr) || !errors.As(err, &appErr) || appErr.Code != errcode.RedisError {
		t.Fatalf("Invalidate() = %v; want Redis error", err)
	}
}
