// Package mq 把判题相关的消息收发(工作队列 + 单题进度发布/订阅)收敛到一处,
// 对外暴露语义化方法,底层用 Redis 实现(Stream 消费者组做工作队列、pub/sub 做进度),
// 调用方不必关心 key 命名和 Redis 细节。
package mq

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Redis key / channel 命名(原 pkg/rediskey,收进 mq 内部,不再外泄)
const (
	submissionStream = "submission_stream" // 工作队列(Redis Stream)
	submissionGroup  = "judge_workers"     // 消费者组
	submissionField  = "sid"               // Stream 消息里存提交 id 的字段名
	versionField     = "version"           // Submission 重判版本，和 sid 一起唯一标识一次评测
	submissionDLQ    = "submission_dlq"    // 死信队列(仍用 list,仅供人工排查)
)

func updateChannel(submissionID int64) string {
	return fmt.Sprintf("submission_update_%d", submissionID)
}

func doneChannel(submissionID int64) string {
	return fmt.Sprintf("submission_done_%d", submissionID)
}

// MQ 判题消息收发器,底层是 Redis。
type MQ struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *MQ {
	m := &MQ{rdb: rdb}
	m.ensureGroup(context.Background())
	return m
}

// StreamTask 从工作队列读到的一条任务。
// Versioned=false 表示旧版本生产者只写入了 sid；这类消息仅兼容初始版本（version=0）。
type StreamTask struct {
	SubmissionID int64
	Version      int
	Versioned    bool
	MsgID        string
}

// ==================== 判题工作队列(Redis Stream + 消费者组) ====================

// ensureGroup 幂等创建 Stream 与消费者组(MKSTREAM 顺带建流);组已存在返回 BUSYGROUP,忽略。
func (m *MQ) ensureGroup(ctx context.Context) {
	err := m.rdb.XGroupCreateMkStream(ctx, submissionStream, submissionGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		// 建组失败不致命：XADD 仍可用；消费端会在 NOGROUP 时重试建组
		_ = err
	}
}

// EnqueueSubmission 把待评测的 Submission 版本入队(XADD)。
func (m *MQ) EnqueueSubmission(ctx context.Context, submissionID int64, version int) error {
	if submissionID <= 0 {
		return fmt.Errorf("submission id must be positive: %d", submissionID)
	}
	if version < 0 {
		return fmt.Errorf("submission version must not be negative: %d", version)
	}
	return m.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: submissionStream,
		Values: map[string]any{
			submissionField: submissionID,
			versionField:    version,
		},
	}).Err()
}

// ReadNewSubmission 阻塞读一条新任务(">"),block 时间内没有返回 (nil, nil)。
func (m *MQ) ReadNewSubmission(ctx context.Context, consumer string, block time.Duration) (*StreamTask, error) {
	res, err := m.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    submissionGroup,
		Consumer: consumer,
		Streams:  []string{submissionStream, ">"},
		Count:    1,
		Block:    block,
	}).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil // block 超时,无新消息
	}
	if err != nil {
		// 组不存在(如 Redis 被清过)时补建后交由上层重试
		if strings.Contains(err.Error(), "NOGROUP") {
			m.ensureGroup(ctx)
			return nil, nil
		}
		return nil, err
	}
	return m.firstTask(ctx, res)
}

// ReadPendingSubmissions 取本消费者上次残留、尚未 ACK 的任务(ID "0"),用于重启后重投,避免丢单。
func (m *MQ) ReadPendingSubmissions(ctx context.Context, consumer string, count int64) ([]StreamTask, error) {
	res, err := m.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    submissionGroup,
		Consumer: consumer,
		Streams:  []string{submissionStream, "0"},
		Count:    count,
	}).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "NOGROUP") {
			m.ensureGroup(ctx)
			return nil, nil
		}
		return nil, err
	}
	var tasks []StreamTask
	for _, s := range res {
		for _, msg := range s.Messages {
			if t := toTask(msg); t != nil {
				tasks = append(tasks, *t)
				continue
			}
			if err := m.AckSubmission(ctx, msg.ID); err != nil {
				return nil, fmt.Errorf("discard invalid pending submission message %s: %w", msg.ID, err)
			}
		}
	}
	return tasks, nil
}

// AckSubmission 确认并删除已处理完的消息(XACK + XDEL,避免 Stream 无限增长)。
func (m *MQ) AckSubmission(ctx context.Context, msgID string) error {
	if err := m.rdb.XAck(ctx, submissionStream, submissionGroup, msgID).Err(); err != nil {
		return err
	}
	return m.rdb.XDel(ctx, submissionStream, msgID).Err()
}

// firstTask 从 XReadGroup 结果里取第一条；无法解析的毒消息会被确认并删除，避免永久占用 PEL。
func (m *MQ) firstTask(ctx context.Context, res []goredis.XStream) (*StreamTask, error) {
	for _, s := range res {
		for _, msg := range s.Messages {
			if t := toTask(msg); t != nil {
				return t, nil
			}
			if err := m.AckSubmission(ctx, msg.ID); err != nil {
				return nil, fmt.Errorf("discard invalid submission message %s: %w", msg.ID, err)
			}
			return nil, nil
		}
	}
	return nil, nil
}

// toTask 把一条 Stream 消息解析成 StreamTask；字段缺失或非法时返回 nil，由读取方 ACK 丢弃。
func toTask(msg goredis.XMessage) *StreamTask {
	raw, ok := msg.Values[submissionField]
	if !ok {
		return nil
	}
	sid, err := strconv.ParseInt(fmt.Sprint(raw), 10, 64)
	if err != nil || sid <= 0 {
		return nil
	}

	task := &StreamTask{SubmissionID: sid, MsgID: msg.ID}
	rawVersion, ok := msg.Values[versionField]
	if !ok {
		return task
	}
	version, err := strconv.Atoi(fmt.Sprint(rawVersion))
	if err != nil || version < 0 {
		return nil
	}
	task.Version = version
	task.Versioned = true
	return task
}

// DeadLetterSubmission 把无法处理的任务投入死信队列。
func (m *MQ) DeadLetterSubmission(ctx context.Context, payload string) error {
	return m.rdb.RPush(ctx, submissionDLQ, payload).Err()
}

// ==================== 单题评测进度 ====================

// PublishUpdate 发布某提交的中间进度(每个测试点)。
func (m *MQ) PublishUpdate(ctx context.Context, submissionID int64, payload string) error {
	return m.rdb.Publish(ctx, updateChannel(submissionID), payload).Err()
}

// PublishDone 发布某提交的最终结果。
func (m *MQ) PublishDone(ctx context.Context, submissionID int64, payload string) error {
	return m.rdb.Publish(ctx, doneChannel(submissionID), payload).Err()
}

// SubmissionEvent 订阅端拿到的事件,Done=true 表示评测结束。
type SubmissionEvent struct {
	Done    bool
	Payload string
}

// SubscribeSubmission 订阅某提交的进度与完成事件(供 SSE 推送)。
// 返回事件 channel 和清理函数,调用方需 defer cleanup()。
func (m *MQ) SubscribeSubmission(ctx context.Context, submissionID int64) (<-chan SubmissionEvent, func()) {
	done := doneChannel(submissionID)
	sub := m.rdb.Subscribe(ctx, updateChannel(submissionID), done)
	raw := sub.Channel()
	out := make(chan SubmissionEvent)
	go func() {
		defer close(out)
		for msg := range raw {
			select {
			case out <- SubmissionEvent{Done: msg.Channel == done, Payload: msg.Payload}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, func() { sub.Close() }
}
