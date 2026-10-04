package service

import (
	"context"
	"errors"
	"strings"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

type LanguageService struct {
	repo *repository.LanguageRepo
}

type LanguageResolver interface {
	ResolveEnabledName(ctx context.Context, id int) (string, error)
}

// SaveLanguageParams 创建/更新语言的输入。
type SaveLanguageParams struct {
	Name   string
	Status int
	Sort   int
}

func NewLanguageService(repo *repository.LanguageRepo) *LanguageService {
	return &LanguageService{repo: repo}
}

func (s *LanguageService) ListEnabled(ctx context.Context) ([]model.Language, error) {
	return s.list(ctx, true)
}

func (s *LanguageService) AdminList(ctx context.Context) ([]model.Language, error) {
	return s.list(ctx, false)
}

func (s *LanguageService) list(ctx context.Context, onlyEnabled bool) ([]model.Language, error) {
	list, err := s.repo.List(ctx, onlyEnabled)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return list, nil
}

// ResolveEnabledName 将前端语言编号解析为提交记录使用的稳定名称。
func (s *LanguageService) ResolveEnabledName(ctx context.Context, id int) (string, error) {
	language, err := s.repo.GetEnabledByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errcode.ErrUnsupportedLanguage
		}
		return "", errcode.ErrDatabase.Wrap(err)
	}
	if !judge.SupportsLanguage(language.Name) {
		return "", errcode.ErrUnsupportedLanguage.WithMsg("该语言尚未配置评测预设")
	}
	return language.Name, nil
}

func (s *LanguageService) Create(ctx context.Context, params SaveLanguageParams) (*model.Language, error) {
	name, err := validateLanguageName(params.Name)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsByName(ctx, name, 0)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if exists {
		return nil, errcode.ErrInvalidParams.WithMsg("编程语言已存在")
	}
	language := &model.Language{Name: name, Status: params.Status, Sort: params.Sort}
	if err := s.repo.Create(ctx, language); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return language, nil
}

func (s *LanguageService) Update(ctx context.Context, id int, params SaveLanguageParams) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInvalidParams.WithMsg("编程语言不存在")
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	name, err := validateLanguageName(params.Name)
	if err != nil {
		return err
	}
	exists, err := s.repo.ExistsByName(ctx, name, id)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	if exists {
		return errcode.ErrInvalidParams.WithMsg("编程语言已存在")
	}
	if err := s.repo.Update(ctx, id, map[string]any{"name": name, "status": params.Status, "sort": params.Sort}); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *LanguageService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInvalidParams.WithMsg("编程语言不存在")
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func validateLanguageName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errcode.ErrInvalidParams.WithMsg("请输入语言名称")
	}
	canonical, ok := judge.CanonicalLanguageName(name)
	if !ok {
		return "", errcode.ErrUnsupportedLanguage.WithMsg("评测程序暂不支持该语言")
	}
	return canonical, nil
}
