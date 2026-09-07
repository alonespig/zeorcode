package judgeworker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"
	"zoj/pkg/remoteoj"
	"zoj/pkg/util"

	"go.uber.org/zap"
)

const (
	remotePollInterval = 2 * time.Second // 轮询间隔
	remoteMaxPolls     = 150             // 最多轮询次数(≈5min)，超时按未知错误处理
)

// pendingRemote 一个待轮询的远程提交
type pendingRemote struct {
	sub       model.Submission
	task      mq.StreamTask
	oj        remoteoj.RemoteOJ
	acc       *remoteoj.RemoteAccount
	remoteID  string
	accountID int64 // 占用的账号池账号，出结果后归还
	polls     int
	persisted bool
}

// dispatchRemote 在 worker 协程里把提交推到远程 OJ，拿到远程提交号后交给轮询协程。
// 从账号池抢占一个账号(独占至判完)，Submit 较快，长耗时的判题等待由轮询协程异步承担。
func (w *Worker) dispatchRemote(
	ctx context.Context,
	task mq.StreamTask,
	s model.Submission,
	problem *model.Problem,
	log *zap.SugaredLogger,
) (taskProcessOutcome, error) {
	oj, ok := remoteoj.Get(problem.OJ)
	if !ok {
		log.Errorw("unsupported remote oj", "oj", problem.OJ)
		return w.completeFailedRemote(ctx, &s, log)
	}

	m, err := w.accRepo.AcquireByOJ(ctx, problem.OJ)
	if err != nil {
		log.Errorw("acquire remote account failed", "oj", problem.OJ, "err", err)
		return w.completeFailedRemote(ctx, &s, log)
	}
	if m == nil {
		// 区分"压根没配账号"(直接失败，否则会无限回队列)和"账号都忙"(延迟重试)
		if cnt, _ := w.accRepo.CountEnabledByOJ(ctx, problem.OJ); cnt == 0 {
			log.Errorw("no remote account configured", "oj", problem.OJ)
			return w.completeFailedRemote(ctx, &s, log)
		}
		log.Infow("all remote accounts busy, requeue", "oj", problem.OJ)
		return w.requeueRemote(ctx, task)
	}

	acc := buildRemoteAccount(m)
	remoteID, err := oj.Submit(acc, remoteoj.SubmitReq{
		RemoteProblemID: problem.RemoteProblemID,
		Language:        s.Language,
		Code:            s.Code,
	})
	if err != nil {
		log.Warnw("remote submit failed", "oj", problem.OJ, "err", err)
		w.accRepo.Release(ctx, m.ID) // 提交失败立即归还账号
		w.sendToDLQ(s.ID, dlqStageRemoteSubmit, err)
		return w.completeFailedRemote(ctx, &s, log)
	}

	log.Infow("remote submitted", "oj", problem.OJ, "remoteID", remoteID, "accountID", m.ID)
	select {
	case w.remoteTasks <- &pendingRemote{sub: s, task: task, oj: oj, acc: acc, remoteID: remoteID, accountID: m.ID}:
		return taskProcessDeferred, nil
	case <-ctx.Done():
		w.accRepo.Release(context.Background(), m.ID)
		return taskProcessDeferred, ctx.Err()
	}
}

func (w *Worker) completeFailedRemote(ctx context.Context, s *model.Submission, log *zap.SugaredLogger) (taskProcessOutcome, error) {
	persisted, err := w.failRemote(ctx, s, log)
	if err != nil {
		return taskProcessDeferred, err
	}
	if !persisted {
		return taskProcessDeferred, nil
	}
	return taskProcessCompleted, nil
}

// requeueRemote 等待一小段时间后释放当前认领并写入新消息；新消息成功入队后旧消息才可 ACK。
func (w *Worker) requeueRemote(ctx context.Context, task mq.StreamTask) (taskProcessOutcome, error) {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return taskProcessDeferred, ctx.Err()
	case <-timer.C:
	}

	dispatch := repository.SubmissionDispatch{SubmissionID: task.SubmissionID, Version: task.Version}
	released, err := w.subRepo.ReleaseDispatchForRetry(ctx, dispatch, task.MsgID)
	if err != nil {
		return taskProcessDeferred, err
	}
	if !released {
		return taskProcessDeferred, nil
	}
	if err := w.mq.EnqueueSubmission(ctx, task.SubmissionID, task.Version); err != nil {
		logger.Errorw("requeue remote submission failed", "submissionID", task.SubmissionID, "version", task.Version, "err", err)
		return taskProcessDeferred, err
	}
	return taskProcessRequeued, nil
}

// remotePollLoop 单协程：收新任务入列，每个 tick 轮询所有在判的远程提交，出结果就落库。
func (w *Worker) remotePollLoop(ctx context.Context) {
	ticker := time.NewTicker(remotePollInterval)
	defer ticker.Stop()

	var pending []*pendingRemote

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-w.remoteTasks:
			pending = append(pending, p)
		case <-ticker.C:
			pending = w.pollPending(ctx, pending)
		}
	}
}

// pollPending 轮询一轮，返回仍未出结果的列表
func (w *Worker) pollPending(ctx context.Context, pending []*pendingRemote) []*pendingRemote {
	kept := pending[:0]
	for _, p := range pending {
		log := logger.S().With("submissionID", p.sub.ID, "remoteID", p.remoteID)
		if p.persisted {
			if !w.finishRemoteTask(ctx, p, log) {
				kept = append(kept, p)
			}
			continue
		}

		p.polls++

		v, err := p.oj.Poll(p.acc, p.remoteID)
		if err != nil {
			if p.polls >= remoteMaxPolls {
				log.Warnw("remote poll gave up after retries", "err", err)
				p.persisted, _ = w.failRemote(ctx, &p.sub, log)
				if p.persisted {
					w.accRepo.Release(ctx, p.accountID)
					if !w.finishRemoteTask(ctx, p, log) {
						kept = append(kept, p)
					}
					continue
				}
			}
			log.Debugw("remote poll error, will retry", "err", err)
			kept = append(kept, p)
			continue
		}

		if v.Status == judge.Pending {
			if p.polls >= remoteMaxPolls {
				log.Warnw("remote judge timeout")
				p.persisted, _ = w.failRemote(ctx, &p.sub, log)
				if p.persisted {
					w.accRepo.Release(ctx, p.accountID)
					if !w.finishRemoteTask(ctx, p, log) {
						kept = append(kept, p)
					}
					continue
				}
			}
			kept = append(kept, p)
			continue
		}

		// 出终态：落库 + 推 SSE + 维护统计 + 归还账号
		p.persisted, _ = w.completeRemote(ctx, &p.sub, v, log)
		if !p.persisted {
			kept = append(kept, p)
			continue
		}
		w.accRepo.Release(ctx, p.accountID)
		if !w.finishRemoteTask(ctx, p, log) {
			kept = append(kept, p)
		}
	}
	return kept
}

func (w *Worker) finishRemoteTask(ctx context.Context, p *pendingRemote, log *zap.SugaredLogger) bool {
	dispatch := repository.SubmissionDispatch{SubmissionID: p.task.SubmissionID, Version: p.task.Version}
	completed, err := w.subRepo.CompleteDispatchForJudging(ctx, dispatch, p.task.MsgID)
	if err != nil || !completed {
		log.Errorw("complete remote dispatch failed", "completed", completed, "err", err)
		return false
	}
	if err := w.mq.AckSubmission(ctx, p.task.MsgID); err != nil {
		log.Errorw("ack remote submission failed", "msgID", p.task.MsgID, "err", err)
		return false
	}
	return true
}

// completeRemote 远程出结果后的收尾，对齐本地 finalizeSubmission 的落库/推送/统计
func (w *Worker) completeRemote(ctx context.Context, s *model.Submission, v *remoteoj.Verdict, log *zap.SugaredLogger) (bool, error) {
	s.Status = v.Status
	s.TimeUsed = v.TimeMs
	s.MemoryUsed = v.MemoryKB
	// 远程 OJ 只回一个最终状态、没有测试点明细，无法算测试点均分，
	// 所以作业与比赛中的远程题只能按「过了给满分、否则 0」全有或全无地计分。
	if s.HomeworkID != 0 || s.ContestID != 0 {
		fullScore := model.HomeworkProblemFullScore
		if s.ContestID != 0 {
			fullScore = w.contestProblemFullScore(s.ContestID, s.ProblemID)
		}
		if v.Status == judge.Accepted {
			s.Score = fullScore
		} else {
			s.Score = 0
		}
	}

	if err := w.saveResultGuarded(ctx, s, dlqStageFinalSave, log); err != nil {
		if errors.Is(err, errSubmissionSuperseded) {
			return true, nil
		}
		return false, err
	}

	w.publishRemoteDone(s, log)

	if s.ContestID == 0 {
		w.updateUserProblem(*s, log)
	} else {
		w.updateContestStats(*s, log)
	}
	log.Infow("remote judge done", "status", s.Status, "timeMs", s.TimeUsed, "memoryKB", s.MemoryUsed)
	return true, nil
}

// failRemote 远程提交/判题失败：标未知错误，落库并推送，避免前端一直转圈
func (w *Worker) failRemote(ctx context.Context, s *model.Submission, log *zap.SugaredLogger) (bool, error) {
	s.Status = judge.UnknownError
	s.TimeUsed = 0
	s.MemoryUsed = 0
	if err := w.saveResultGuarded(ctx, s, dlqStageFinalSave, log); err != nil {
		if errors.Is(err, errSubmissionSuperseded) {
			return true, nil
		}
		return false, err
	}
	w.publishRemoteDone(s, log)
	return true, nil
}

// publishRemoteDone 发送评测完成事件(远程题没有逐点结果，caseResults 为空)
func (w *Worker) publishRemoteDone(s *model.Submission, log *zap.SugaredLogger) {
	mp := map[string]any{
		"submission": dto.SubmissionEventInfo{
			ID:         s.PublicID,
			Language:   s.Language,
			Status:     s.Status,
			TimeUsed:   s.TimeUsed,
			MemoryUsed: s.MemoryUsed,
			CreatedAt:  s.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"caseResults": []dto.SubmissionCaseResult{},
	}
	data, err := json.Marshal(mp)
	if err != nil {
		log.Errorw("marshal remote done failed", "err", err)
		return
	}
	if err := w.mq.PublishDone(context.Background(), s.ID, string(data)); err != nil {
		log.Errorw("publish remote done failed", "err", err)
	}
}

// buildRemoteAccount 把账号池取出的记录解密成可用账号(cookie 或 用户名/密码)。
// 解密失败返回的账号 secret 为空，由各 OJ 实现自行报错。
func buildRemoteAccount(m *model.RemoteAccount) *remoteoj.RemoteAccount {
	acc := &remoteoj.RemoteAccount{ID: m.ID, OJ: m.OJ, Username: m.Username, Valid: m.Valid}
	secret, err := util.Decrypt(m.Secret)
	if err != nil {
		logger.Warnw("decrypt remote account secret failed", "oj", m.OJ, "accountID", m.ID, "err", err)
		return acc
	}
	if m.AuthType == "cookie" {
		acc.Cookie = secret
	} else {
		acc.Password = secret
	}
	return acc
}
