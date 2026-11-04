package judgenode

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	judgev1 "github.com/alonespig/zeorcode/judge-node/api/judge/v1"
	"github.com/alonespig/zeorcode/judge-node/internal/artifact"
	judgeapi "github.com/alonespig/zeorcode/judge-node/internal/gojudge"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const AgentVersion = "0.1.0"

type Server struct {
	judgev1.UnimplementedJudgeNodeServiceServer
	nodeID       string
	goJudgeURL   string
	client       *judgeapi.Client
	cache        *ArtifactCache
	semaphore    chan struct{}
	runningTasks atomic.Int32
}

func NewServer(nodeID, goJudgeURL string, cache *ArtifactCache, maxConcurrency int) (*Server, error) {
	if nodeID == "" {
		return nil, errors.New("judge node id is required")
	}
	if goJudgeURL == "" {
		return nil, errors.New("go-judge URL is required")
	}
	if cache == nil {
		return nil, errors.New("artifact cache is required")
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	return &Server{
		nodeID:     nodeID,
		goJudgeURL: goJudgeURL,
		client:     judgeapi.NewClient(goJudgeURL),
		cache:      cache,
		semaphore:  make(chan struct{}, maxConcurrency),
	}, nil
}

func (s *Server) Status(_ context.Context, _ *judgev1.StatusRequest) (*judgev1.StatusResponse, error) {
	probe := judgeapi.Probe(s.goJudgeURL)
	if !probe.Online {
		return nil, status.Error(codes.Unavailable, "go-judge is unavailable")
	}
	return &judgev1.StatusResponse{
		NodeId:         s.nodeID,
		AgentVersion:   AgentVersion,
		GoJudgeVersion: probe.Version,
		MaxConcurrency: int32(cap(s.semaphore)),
		RunningTasks:   s.runningTasks.Load(),
		CacheBytes:     directorySize(s.cache.root),
	}, nil
}

func (s *Server) Judge(req *judgev1.JudgeRequest, stream judgev1.JudgeNodeService_JudgeServer) error {
	if err := validateJudgeRequest(req); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	ctx := stream.Context()
	select {
	case s.semaphore <- struct{}{}:
		s.runningTasks.Add(1)
		defer func() {
			s.runningTasks.Add(-1)
			<-s.semaphore
		}()
	case <-ctx.Done():
		return status.FromContextError(ctx.Err()).Err()
	}

	want := artifact.Descriptor{
		ProblemID: req.Problem.ProblemId,
		Version:   req.Problem.Version,
		Digest:    req.Problem.Digest,
	}
	if err := sendEvent(stream, judgev1.JudgeEventType_JUDGE_EVENT_TYPE_SYNCING, "syncing problem data", nil, nil); err != nil {
		return err
	}
	dir, info, err := s.cache.Ensure(ctx, want)
	if err != nil {
		return status.Errorf(codes.Unavailable, "sync problem data: %v", err)
	}
	if err := sendEvent(stream, judgev1.JudgeEventType_JUDGE_EVENT_TYPE_COMPILING, "compiling", nil, nil); err != nil {
		return err
	}
	artifact, err := s.client.Compile(req.Language, req.SourceCode)
	if err != nil {
		if judgeapi.IsUnavailable(err) {
			return status.Error(codes.Unavailable, "go-judge compile request failed")
		}
		final := &judgev1.FinalResult{
			Status:        int32(judgeapi.CompileError),
			CompileOutput: judgeapi.CompilationOutput(err),
		}
		return sendEvent(stream, judgev1.JudgeEventType_JUDGE_EVENT_TYPE_FINISHED, "compile failed", nil, final)
	}
	defer func() { _ = s.client.DeleteArtifact(artifact) }()

	limits := judgeapi.Limits{
		TimeMs:      int(req.TimeLimitMs),
		MemoryMB:    int(req.MemoryLimitMb),
		OutputBytes: int(req.OutputLimitBytes),
	}
	finalStatus := judgeapi.Accepted
	results := make([]*judgev1.CaseResult, 0, len(info.Cases))
	for _, tc := range info.Cases {
		if err := ctx.Err(); err != nil {
			return status.FromContextError(err).Err()
		}
		if filepath.Base(tc.Input) != tc.Input {
			return status.Error(codes.DataLoss, "invalid cached testcase path")
		}
		input, err := os.ReadFile(filepath.Join(dir, tc.Input))
		if err != nil {
			return status.Error(codes.DataLoss, "cached testcase input is missing")
		}
		verdict, err := s.client.Run(artifact, string(input), limits)
		if err != nil {
			if judgeapi.IsUnavailable(err) {
				return status.Error(codes.Unavailable, "go-judge run request failed")
			}
			return status.Error(codes.Internal, "judge testcase failed")
		}
		caseStatus := verdict.Status
		if caseStatus == judgeapi.Accepted {
			caseStatus = compareOutput(verdict.Stdout, tc.StrippedMd5, tc.AllStrippedMd5)
		}
		if finalStatus == judgeapi.Accepted && caseStatus != judgeapi.Accepted {
			finalStatus = caseStatus
		}
		result := &judgev1.CaseResult{
			CaseId:      int32(tc.ID),
			Status:      int32(caseStatus),
			TimeNs:      verdict.TimeNs,
			MemoryBytes: verdict.MemoryByte,
		}
		results = append(results, result)
		if err := sendEvent(stream, judgev1.JudgeEventType_JUDGE_EVENT_TYPE_CASE_RESULT, "", result, nil); err != nil {
			return err
		}
	}
	final := &judgev1.FinalResult{Status: int32(finalStatus), Cases: results}
	return sendEvent(stream, judgev1.JudgeEventType_JUDGE_EVENT_TYPE_FINISHED, "finished", nil, final)
}

func validateJudgeRequest(req *judgev1.JudgeRequest) error {
	if req == nil || req.Problem == nil {
		return errors.New("problem artifact is required")
	}
	if req.SubmissionId <= 0 || req.SubmissionVersion < 0 {
		return errors.New("invalid submission identity")
	}
	if req.Language == "" || req.SourceCode == "" {
		return errors.New("language and source code are required")
	}
	if req.Problem.ProblemId <= 0 || req.Problem.Version == "" || !validProblemDigest(req.Problem.Digest) {
		return errors.New("invalid problem artifact")
	}
	if req.TimeLimitMs <= 0 || req.MemoryLimitMb <= 0 {
		return errors.New("invalid resource limits")
	}
	return nil
}

func validProblemDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sendEvent(stream judgev1.JudgeNodeService_JudgeServer, eventType judgev1.JudgeEventType, message string, result *judgev1.CaseResult, final *judgev1.FinalResult) error {
	return stream.Send(&judgev1.JudgeEvent{Type: eventType, Message: message, CaseResult: result, FinalResult: final})
}

func compareOutput(stdout, strippedMD5, allStrippedMD5 string) int {
	if judgeapi.MD5(judgeapi.RtrimOutput(stdout)) == strippedMD5 {
		return judgeapi.Accepted
	}
	if judgeapi.MD5(judgeapi.StripAllSpace(stdout)) == allStrippedMD5 {
		return judgeapi.PresentationError
	}
	return judgeapi.WrongAnswer
}

func directorySize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func (s *Server) String() string {
	return fmt.Sprintf("judge-node(%s)", s.nodeID)
}
