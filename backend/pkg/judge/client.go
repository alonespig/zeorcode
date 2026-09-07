package judge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CompilationFailure 描述用户代码未通过编译，并保留沙箱返回的原始输出。
// 它不是判题机不可用错误，worker 不应因此切换健康实例。
type CompilationFailure struct {
	Language string
	Status   string
	Output   string
}

func (e *CompilationFailure) Error() string {
	if e.Output == "" {
		return fmt.Sprintf("compile failed, language=%s, status=%s", e.Language, e.Status)
	}
	return fmt.Sprintf("compile failed, language=%s, status=%s, output=%s", e.Language, e.Status, e.Output)
}

// CompilationOutput 从编译失败错误链中提取可展示给提交者的原始编译器输出。
func CompilationOutput(err error) string {
	var failure *CompilationFailure
	if !errors.As(err, &failure) {
		return ""
	}
	return failure.Output
}

// Client 判题机客户端，负责与沙箱 (go-judge) 交互
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 用沙箱地址构造客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Compile 编译用户代码，成功后返回可复用的 Artifact
func (c *Client) Compile(language, code string) (*Artifact, error) {
	lang := NormalizeLanguage(language)
	spec, ok := languages[lang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	cmd, cacheName := spec.compile(code)

	results, err := c.runCommand([]map[string]any{cmd})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("compile request returned empty result")
	}

	r := results[0]
	fileID := r.FileIds[cacheName]
	if r.Status != "Accepted" || fileID == "" {
		output := strings.TrimSpace(r.Files.Stderr)
		if output == "" {
			output = strings.TrimSpace(r.Files.Stdout)
		}
		return nil, &CompilationFailure{Language: lang, Status: r.Status, Output: output}
	}

	return &Artifact{Language: lang, FileID: fileID}, nil
}

// Run 用 Artifact 跑一组输入，返回运行结果（含 stdout）。
// 输出对错(AC/PE/WA)不在此判定，由 worker 按测试数据清单(info.json)的哈希比对。
func (c *Client) Run(art *Artifact, input string, lim Limits) (*Verdict, error) {
	if art == nil {
		return nil, errors.New("artifact is nil")
	}
	if art.FileID == "" {
		return nil, errors.New("artifact file id is empty")
	}
	spec, ok := languages[art.Language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", art.Language)
	}

	if lim.TimeMs <= 0 {
		lim.TimeMs = 1000
	}
	if lim.MemoryMB <= 0 {
		lim.MemoryMB = 128
	}

	args, runFile, procLimit := spec.run(lim.MemoryMB)
	cpuLimitNs := int64(lim.TimeMs) * 1_000_000
	clockLimitNs := cpuLimitNs * 2
	memoryLimitBytes := int64(lim.MemoryMB) * 1024 * 1024

	cmd := map[string]any{
		"args":        args,
		"env":         []string{"PATH=/usr/bin:/bin", "LANG=C.UTF-8"},
		"files":       []map[string]any{{"content": input}, {"name": "stdout", "max": 40 * 1024 * 1024}, {"name": "stderr", "max": 10240}},
		"cpuLimit":    cpuLimitNs,
		"clockLimit":  clockLimitNs,
		"memoryLimit": memoryLimitBytes,
		"procLimit":   procLimit,
		"copyIn": map[string]map[string]string{
			runFile: {"fileId": art.FileID},
		},
	}

	results, err := c.runCommand([]map[string]any{cmd})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("run request returned empty result")
	}

	r := results[0]
	timeNs := r.RunTime
	if timeNs <= 0 {
		timeNs = r.Time
	}

	return &Verdict{
		Status:     inferStatus(&r, lim),
		Stdout:     r.Files.Stdout,
		Stderr:     r.Files.Stderr,
		TimeNs:     timeNs,
		MemoryByte: r.Memory,
	}, nil
}

// DeleteArtifact 释放沙箱缓存的编译产物
func (c *Client) DeleteArtifact(art *Artifact) error {
	if art == nil || art.FileID == "" {
		return nil
	}

	u := fmt.Sprintf("%s/file/%s", c.baseURL, url.PathEscape(art.FileID))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete cached file failed, status=%d, body=%s", resp.StatusCode, string(body))
}

func (c *Client) runCommand(cmd []map[string]any) ([]runResult, error) {
	payload, err := json.Marshal(map[string]any{"cmd": cmd})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/run", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("judge request failed, status=%d, body=%s", resp.StatusCode, string(body))
	}

	var results []runResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// inferStatus 把沙箱原始运行结果折成业务状态码（只判运行态：TLE/MLE/RE/正常跑完）。
// 正常跑完返回 Accepted，表示"运行 OK"，最终 AC/PE/WA 由 worker 按哈希判定。
func inferStatus(r *runResult, lim Limits) int {
	status := strings.ToLower(strings.TrimSpace(r.Status))
	timeNs := r.RunTime
	if timeNs <= 0 {
		timeNs = r.Time
	}

	// 资源用量兜底，避免沙箱 status 字段不稳定时漏判
	if lim.TimeMs > 0 && timeNs > int64(lim.TimeMs)*1_000_000 {
		return TimeLimitExceeded
	}
	if lim.MemoryMB > 0 && r.Memory > int64(lim.MemoryMB)*1024*1024 {
		return MemoryLimitExceeded
	}

	switch {
	case strings.Contains(status, "memory limit"):
		return MemoryLimitExceeded
	case strings.Contains(status, "time limit"):
		return TimeLimitExceeded
	case strings.Contains(status, "internal error"), strings.Contains(status, "file error"):
		return UnknownError
	case strings.Contains(status, "nonzero exit"),
		strings.Contains(status, "signalled"),
		strings.Contains(status, "dangerous syscall"),
		strings.Contains(status, "output limit"):
		return RuntimeError
	case strings.Contains(status, "accepted"):
		return Accepted
	default:
		if r.ExitStatus != 0 {
			return RuntimeError
		}
		return Accepted
	}
}
