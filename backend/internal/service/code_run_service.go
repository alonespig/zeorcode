package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/infra/cache"
	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	judgeapi "github.com/alonespig/zeorcode/backend/pkg/judge"

	"gorm.io/gorm"
)

const (
	maxCodeRunCodeBytes   = 128 * 1024
	maxCodeRunInputBytes  = 1024 * 1024
	maxCodeRunOutputBytes = 1024 * 1024
	codeRunInterval       = 2 * time.Second
)

type CodeRunPool interface {
	Size() int
	Next() *judgeapi.Client
	MarkDown(*judgeapi.Client)
}

type CodeRunService struct {
	problems  SubmissionProblemReader
	languages LanguageResolver
	cache     SubmissionCache
	pool      CodeRunPool
}

func NewCodeRunService(
	problems SubmissionProblemReader,
	languages LanguageResolver,
	cache SubmissionCache,
	pool CodeRunPool,
) *CodeRunService {
	return &CodeRunService{problems: problems, languages: languages, cache: cache, pool: pool}
}

func (s *CodeRunService) Run(ctx context.Context, params CodeRunParams, userID int64, isAdmin bool) (*CodeRunResult, error) {
	if err := validateCodeRunParams(params); err != nil {
		return nil, err
	}
	if err := codeRunRateLimit(ctx, s.cache, userID); err != nil {
		return nil, err
	}

	problem, err := s.problems.GetByDisplayID(ctx, params.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if problem.Hidden && !isAdmin {
		return nil, errcode.ErrProblemNotFound
	}

	language, err := s.languages.ResolveEnabledName(ctx, params.Language)
	if err != nil {
		return nil, err
	}
	if s.pool == nil || s.pool.Size() == 0 {
		return nil, errcode.ErrJudgeUnavailable.WithMsg("未配置可用的评测机")
	}

	limits := judgeapi.Limits{
		TimeMs:      problem.TimeLimit,
		MemoryMB:    problem.MemoryLimit,
		OutputBytes: maxCodeRunOutputBytes,
	}
	var lastErr error
	for range s.pool.Size() {
		client := s.pool.Next()
		if client == nil {
			break
		}

		artifact, compileErr := client.Compile(language, params.Code)
		if compileErr != nil {
			var compilationFailure *judgeapi.CompilationFailure
			if errors.As(compileErr, &compilationFailure) {
				return &CodeRunResult{
					Status:        judgeapi.CompileError,
					CompileOutput: compilationFailure.Output,
				}, nil
			}
			lastErr = compileErr
			s.pool.MarkDown(client)
			continue
		}

		verdict, runErr := client.Run(artifact, params.Input, limits)
		cleanupErr := client.DeleteArtifact(artifact)
		if cleanupErr != nil {
			logger.Warnw("delete custom run artifact failed", "problemID", problem.ID, "err", cleanupErr)
		}
		if runErr != nil {
			lastErr = runErr
			s.pool.MarkDown(client)
			continue
		}

		return &CodeRunResult{
			Status:     verdict.Status,
			TimeUsed:   verdict.TimeNs / 1_000_000,
			MemoryUsed: verdict.MemoryByte / 1024,
			Stdout:     verdict.Stdout,
			Stderr:     verdict.Stderr,
		}, nil
	}

	if lastErr == nil {
		lastErr = errors.New("judge pool returned no client")
	}
	return nil, errcode.ErrJudgeUnavailable.Wrap(lastErr)
}

func validateCodeRunParams(params CodeRunParams) error {
	if strings.TrimSpace(params.ProblemID) == "" {
		return errcode.ErrInvalidParams.WithMsg("题号不能为空")
	}
	if strings.TrimSpace(params.Code) == "" {
		return errcode.ErrInvalidParams.WithMsg("代码不能为空")
	}
	if len(params.Code) > maxCodeRunCodeBytes {
		return errcode.ErrInvalidParams.WithMsg("代码不能超过 128 KB")
	}
	if len(params.Input) > maxCodeRunInputBytes {
		return errcode.ErrInvalidParams.WithMsg("自测输入不能超过 1 MB")
	}
	if params.Language <= 0 {
		return errcode.ErrInvalidParams.WithMsg("请选择编程语言")
	}
	return nil
}

func codeRunRateLimit(ctx context.Context, c SubmissionCache, userID int64) error {
	ok, err := c.SetNX(ctx, cache.CodeRunLock(userID), 1, codeRunInterval)
	if err != nil {
		return errcode.ErrRedis.Wrap(err)
	}
	if !ok {
		return errcode.ErrTooManyRequests.WithMsg("运行过于频繁，请稍后再试")
	}
	return nil
}
