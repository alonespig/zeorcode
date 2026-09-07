package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"

	"gorm.io/gorm"
)

// parseTagIDs 把 "1,2,3" 解析成 []int64（忽略非法/空）
func parseTagIDs(s string) []int64 {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

type ProblemService struct {
	repo     *repository.ProblemRepo
	subRepo  *repository.SubmissionRepo
	testData ProblemTestDataReader
	db       *gorm.DB
}

func NewProblemService(repo *repository.ProblemRepo,
	subRepo *repository.SubmissionRepo,
	testData ProblemTestDataReader,
	db *gorm.DB) *ProblemService {
	return &ProblemService{repo: repo, subRepo: subRepo, testData: testData, db: db}
}

// ResolveID 把对外题号解析成内部主键（handler 用它把 URL 里的题号换成主键）。
func (s *ProblemService) ResolveID(ctx context.Context, displayID string) (int64, error) {
	id, err := s.repo.ResolveID(ctx, displayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrProblemNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return id, nil
}

func (s *ProblemService) Create(ctx context.Context, req *dto.CreateProblemReq) (string, error) {
	// 题号查重
	if exists, err := s.repo.ExistsByDisplayID(ctx, req.DisplayID, 0); err != nil {
		return "", errcode.ErrDatabase.Wrap(err)
	} else if exists {
		return "", errcode.ErrProblemDisplayIDExists
	}
	problem := model.Problem{
		DisplayID:       req.DisplayID,
		Name:            req.Name,
		Difficulty:      req.Difficulty,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		Description:     req.Description,
		InputFormat:     req.InputFormat,
		OutputFormat:    req.OutputFormat,
		Hint:            req.Hint,
		OJ:              req.OJ,
		RemoteProblemID: req.RemoteProblemID,
		Hidden:          req.Hidden,
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&problem).Error; err != nil {
			return err
		}
		if len(req.TagsID) > 0 {
			pts := make([]model.ProblemTag, 0, len(req.TagsID))
			for _, tagID := range req.TagsID {
				pts = append(pts, model.ProblemTag{
					ProblemID: problem.ID,
					TagID:     tagID,
				})
			}
			if err := tx.Create(&pts).Error; err != nil {
				return err
			}
		}
		if len(req.Samples) > 0 {
			samples := make([]model.ProblemSamples, 0, len(req.Samples))
			for _, sample := range req.Samples {
				samples = append(samples, model.ProblemSamples{
					ProblemID: problem.ID,
					Input:     sample.Input,
					Output:    sample.Output,
					Explain:   sample.Explain,
				})
			}
			if err := tx.Create(&samples).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err := os.MkdirAll(fmt.Sprintf(
		"ojdata/problems/%d", problem.ID), 0755); err != nil {
		return "", errcode.ErrInternal.WithMsg("创建题目数据目录失败").Wrap(err)
	}
	// 返回对外题号（前端据此跳转 /problem/{题号}/file）
	return problem.DisplayID, err
}

func (s *ProblemService) List(ctx context.Context, form *dto.ProblemListReq, userID *int64, isAdmin bool) (*dto.ProblemListResp, error) {
	problems, total, err := s.repo.List(ctx, &repository.ProblemQuery{
		Page:          form.Page,
		PageSize:      form.PageSize,
		Keyword:       form.Keyword,
		Order:         form.Order,
		Difficulty:    form.Difficulty,
		TagIDs:        parseTagIDs(form.Tags),
		IncludeHidden: isAdmin, // 管理员才能在列表看到隐藏题
	})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	resp := dto.ProblemListResp{
		Total: int(total),
		List:  make([]dto.ProblemItemResp, 0, len(problems)),
	}

	problemIDs := make([]int64, 0, len(problems))
	for _, problem := range problems {
		problemIDs = append(problemIDs, problem.ID)
	}

	statusMap := map[int64]int{}
	if userID != nil {
		var err error
		statusMap, err = s.repo.GetUserProblemStatusByIDs(ctx, *userID, problemIDs)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
	}

	tagsByProblem := make(map[int64][]dto.TagItem, len(problems))
	rawTags, err := s.repo.GetProblemTagsByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, tag := range rawTags {
		tagsByProblem[tag.ProblemID] = append(tagsByProblem[tag.ProblemID], dto.TagItem{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	statsByProblem := make(map[int64]repository.ProblemSubmissionStat, len(problems))
	stats, err := s.subRepo.GetProblemSubmissionStats(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, stat := range stats {
		statsByProblem[stat.ProblemID] = stat
	}

	for _, problem := range problems {
		tags := tagsByProblem[problem.ID]
		stat := statsByProblem[problem.ID]
		item := dto.ProblemItemResp{
			ID:            problem.DisplayID, // 对外用题号，不暴露自增主键
			Name:          problem.Name,
			Tags:          tags,
			Difficulty:    problem.Difficulty,
			AcceptedCount: int(stat.AcceptedCount),
			SubmitCount:   int(stat.SubmitCount),
			CreatedAt:     problem.CreatedAt.Format("2006-01-02 15:04:05"),
			Hidden:        problem.Hidden,
			OJ:            problem.OJ,
		}
		status, ok := statusMap[problem.ID]
		if ok {
			item.Status = &status
		}
		resp.List = append(resp.List, item)
	}

	return &resp, nil
}

// fetchTagsAndSamples 获取题目的标签和样例（公用逻辑）
func (s *ProblemService) fetchTagsAndSamples(ctx context.Context, problemID int64) ([]dto.TagItem, []dto.ProblemSample, error) {
	rawTags, _ := s.repo.GetProblemTagByID(ctx, problemID)
	tags := make([]dto.TagItem, 0, len(rawTags))
	for _, tag := range rawTags {
		tags = append(tags, dto.TagItem{ID: tag.ID, Name: tag.Name})
	}

	rawSamples, err := s.repo.GetProblemSamplesByID(ctx, problemID)
	if err != nil {
		return nil, nil, errcode.ErrDatabase.Wrap(err)
	}
	samples := make([]dto.ProblemSample, 0, len(rawSamples))
	for _, s := range rawSamples {
		samples = append(samples, dto.ProblemSample{
			Input: s.Input, Output: s.Output, Explain: s.Explain,
		})
	}
	return tags, samples, nil
}

func (s *ProblemService) GetByID(ctx context.Context, id int64, isAdmin bool) (*dto.ProblemDetailResp, error) {
	problem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	// 隐藏题对非管理员一律当作不存在（不泄露存在性）
	if problem.Hidden && !isAdmin {
		return nil, errcode.ErrProblemNotFound
	}
	hasTestData, err := problemTestDataReady(ctx, s.testData, problem)
	if err != nil {
		return nil, err
	}
	result := 1
	acceptedCount, _ := s.subRepo.GetSubmissionCount(ctx, id, &result)
	submitCount, _ := s.subRepo.GetSubmissionCount(ctx, id, nil)

	tags, samples, err := s.fetchTagsAndSamples(ctx, problem.ID)
	if err != nil {
		return nil, err
	}
	return &dto.ProblemDetailResp{
		ID:              problem.DisplayID, // 对外题号
		Name:            problem.Name,
		Difficulty:      problem.Difficulty,
		TimeLimit:       problem.TimeLimit,
		MemoryLimit:     problem.MemoryLimit,
		Description:     problem.Description,
		AcceptedCount:   int(acceptedCount),
		SubmitCount:     int(submitCount),
		InputFormat:     problem.InputFormat,
		OutputFormat:    problem.OutputFormat,
		Hint:            problem.Hint,
		Samples:         samples,
		Tags:            tags,
		OJ:              problem.OJ,
		RemoteProblemID: problem.RemoteProblemID,
		Hidden:          problem.Hidden,
		HasTestData:     hasTestData,
	}, nil
}

func (s *ProblemService) GetForEdit(ctx context.Context, id int64) (*dto.ProblemModel, error) {
	problem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	tags, samples, err := s.fetchTagsAndSamples(ctx, problem.ID)
	if err != nil {
		return nil, err
	}
	return &dto.ProblemModel{
		ID:           problem.DisplayID, // 对外题号
		Name:         problem.Name,
		Difficulty:   problem.Difficulty,
		TimeLimit:    problem.TimeLimit,
		MemoryLimit:  problem.MemoryLimit,
		Description:  problem.Description,
		InputFormat:  problem.InputFormat,
		OutputFormat: problem.OutputFormat,
		Hint:         problem.Hint,
		Samples:      samples,
		Tags:         tags,
		Hidden:       problem.Hidden,
	}, nil
}

func (s *ProblemService) GetTagList(ctx context.Context) (*dto.TagsListResp, error) {
	tags, err := s.repo.GetTagsList(ctx)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.TagsListResp{
		Tags: make([]dto.TagItem, 0, len(tags)),
	}
	for _, tag := range tags {
		resp.Tags = append(resp.Tags, dto.TagItem{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return resp, nil
}

func (s *ProblemService) GetAdminTagList(ctx context.Context) ([]dto.AdminTagItem, error) {
	tags, err := s.repo.GetAdminTagsList(ctx)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result := make([]dto.AdminTagItem, 0, len(tags))
	for _, tag := range tags {
		result = append(result, dto.AdminTagItem{
			ID:              tag.ID,
			Name:            tag.Name,
			ProblemCount:    tag.ProblemCount,
			ProblemSetCount: tag.ProblemSetCount,
		})
	}
	return result, nil
}

func normalizeTagName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errcode.ErrInvalidParams.WithMsg("标签名称不能为空")
	}
	if utf8.RuneCountInString(name) > 64 {
		return "", errcode.ErrInvalidParams.WithMsg("标签名称不能超过 64 个字符")
	}
	return name, nil
}

func (s *ProblemService) CreateTag(ctx context.Context, name string) (*dto.TagItem, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsTagName(ctx, name, 0)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if exists {
		return nil, errcode.ErrInvalidParams.WithMsg("算法标签已存在")
	}
	tag := &model.Tag{Name: name}
	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return &dto.TagItem{ID: tag.ID, Name: tag.Name}, nil
}

func (s *ProblemService) UpdateTag(ctx context.Context, id int64, name string) (*dto.TagItem, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetTagByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidParams.WithMsg("算法标签不存在")
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	exists, err := s.repo.ExistsTagName(ctx, name, id)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if exists {
		return nil, errcode.ErrInvalidParams.WithMsg("算法标签已存在")
	}
	if err := s.repo.UpdateTagName(ctx, id, name); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return &dto.TagItem{ID: id, Name: name}, nil
}

func (s *ProblemService) DeleteTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return errcode.ErrInvalidParams.WithMsg("非法标签 ID")
	}
	if err := s.repo.DeleteTagWithRelations(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInvalidParams.WithMsg("算法标签不存在")
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *ProblemService) Update(ctx context.Context, req *dto.UpdateProblemReq) (string, error) {
	// 题号查重（排除自身，支持改题号）
	if exists, err := s.repo.ExistsByDisplayID(ctx, req.DisplayID, req.ID); err != nil {
		return "", errcode.ErrDatabase.Wrap(err)
	} else if exists {
		return "", errcode.ErrProblemDisplayIDExists
	}
	problem := model.Problem{
		ID:           req.ID,
		DisplayID:    req.DisplayID,
		Name:         req.Name,
		Difficulty:   req.Difficulty,
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
		Description:  req.Description,
		InputFormat:  req.InputFormat,
		OutputFormat: req.OutputFormat,
		Hint:         req.Hint,
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(&problem).Error; err != nil {
			return err
		}
		// Hidden 是 bool，Updates 对零值(false)不写，单独强制更新以支持"取消隐藏"
		if err := tx.Model(&model.Problem{}).Where("id = ?", problem.ID).
			Update("hidden", req.Hidden).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_id = ?", problem.ID).Delete(&model.ProblemTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_id = ?", problem.ID).Delete(&model.ProblemSamples{}).Error; err != nil {
			return err
		}
		if len(req.Tags) > 0 {
			pts := make([]model.ProblemTag, 0, len(req.Tags))
			for _, tagID := range req.Tags {
				pts = append(pts, model.ProblemTag{
					ProblemID: problem.ID,
					TagID:     tagID,
				})
			}
			if err := tx.Create(&pts).Error; err != nil {
				return err
			}
		}
		if len(req.Samples) > 0 {
			samples := make([]model.ProblemSamples, 0, len(req.Samples))
			for _, sample := range req.Samples {
				samples = append(samples, model.ProblemSamples{
					ProblemID: problem.ID,
					Input:     sample.Input,
					Output:    sample.Output,
					Explain:   sample.Explain,
				})
			}
			if err := tx.Create(&samples).Error; err != nil {
				return err
			}
		}
		return nil
	})
	// 返回对外题号（前端据此跳转 /problem/{题号}/file）
	return problem.DisplayID, err
}
