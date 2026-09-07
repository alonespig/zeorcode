package loj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zoj/pkg/judge"
	"zoj/pkg/remoteoj"
)

const (
	apiBase   = "https://api.loj.ac/api"
	apiURL    = apiBase + "/problem/getProblem"
	loginURL  = apiBase + "/auth/login"
	submitURL = apiBase + "/submission/submit"
	detailURL = apiBase + "/submission/getSubmissionDetail"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// LibreOJ 远程实现
type LibreOJ struct{}

func init() {
	remoteoj.Register(&LibreOJ{})
}

func (l *LibreOJ) Name() string { return "LibreOJ" }

// getProblem 接口返回结构（只取需要的字段）
type lojResp struct {
	LocalizedContents struct {
		Title           string `json:"title"`
		ContentSections []struct {
			SectionTitle string `json:"sectionTitle"`
			Type         string `json:"type"` // Text / Sample
			Text         string `json:"text"`
			SampleID     int    `json:"sampleId"`
		} `json:"contentSections"`
	} `json:"localizedContentsOfLocale"`
	JudgeInfo struct {
		TimeLimit   int `json:"timeLimit"`   // ms
		MemoryLimit int `json:"memoryLimit"` // MB
	} `json:"judgeInfo"`
	Samples []struct {
		InputData  string `json:"inputData"`
		OutputData string `json:"outputData"`
	} `json:"samples"`
}

func (l *LibreOJ) CrawlProblem(_ *remoteoj.RemoteAccount, pid string) (*remoteoj.RemoteProblem, error) {
	displayID, err := strconv.Atoi(strings.TrimSpace(pid))
	if err != nil {
		return nil, fmt.Errorf("LibreOJ 题号必须是数字: %s", pid)
	}

	payload, _ := json.Marshal(map[string]any{
		"displayId":                 displayID,
		"localizedContentsOfLocale": "zh_CN",
		"samples":                   true,
		"judgeInfo":                 true,
	})

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("LibreOJ 返回 %d: %s", resp.StatusCode, string(body))
	}

	var r lojResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("解析 LibreOJ 响应失败: %w", err)
	}
	if r.LocalizedContents.Title == "" {
		return nil, fmt.Errorf("LibreOJ 未找到题目 %s", pid)
	}

	p := &remoteoj.RemoteProblem{
		OJ:              l.Name(),
		RemoteProblemID: pid,
		Title:           r.LocalizedContents.Title,
		TimeLimit:       r.JudgeInfo.TimeLimit,
		MemoryLimit:     r.JudgeInfo.MemoryLimit,
		Source:          fmt.Sprintf("https://loj.ac/p/%s", pid),
	}

	// 样例：来自顶层 samples（输入/输出）
	for _, s := range r.Samples {
		p.Samples = append(p.Samples, remoteoj.ProblemSample{
			Input:  s.InputData,
			Output: s.OutputData,
		})
	}

	// 内容分段：按小标题归类；Sample 段的说明文字附到对应样例
	var descParts []string
	for _, sec := range r.LocalizedContents.ContentSections {
		switch {
		case sec.Type == "Sample":
			if sec.SampleID >= 0 && sec.SampleID < len(p.Samples) {
				p.Samples[sec.SampleID].Explain = sec.Text
			}
		case strings.Contains(sec.SectionTitle, "输入"):
			p.InputFormat = sec.Text
		case strings.Contains(sec.SectionTitle, "输出"):
			p.OutputFormat = sec.Text
		case strings.Contains(sec.SectionTitle, "数据范围"),
			strings.Contains(sec.SectionTitle, "提示"),
			strings.Contains(sec.SectionTitle, "范围"):
			if p.Hint != "" {
				p.Hint += "\n\n"
			}
			p.Hint += sec.Text
		case strings.Contains(sec.SectionTitle, "描述"):
			descParts = append(descParts, sec.Text)
		default:
			// 其它 Text 段落并入描述（带小标题）
			if sec.Type == "Text" && sec.Text != "" {
				descParts = append(descParts, "**"+sec.SectionTitle+"**\n\n"+sec.Text)
			}
		}
	}
	p.Description = strings.Join(descParts, "\n\n")

	return p, nil
}

// lojLang zoj 语言 -> LibreOJ language + compileAndRunOptions
type lojLang struct {
	name string
	opts map[string]any
}

var lojLangMap = map[string]lojLang{
	"cpp":    {"cpp", map[string]any{"compiler": "g++", "std": "c++17", "O": "2", "m": "64"}},
	"java":   {"java", map[string]any{}},
	"python": {"python", map[string]any{"version": "3.10"}},
}

// Submit 账号密码登录拿 token → 用 displayId 换内部 problemId → 提交，返回 submissionId。
func (l *LibreOJ) Submit(acc *remoteoj.RemoteAccount, req remoteoj.SubmitReq) (string, error) {
	if acc == nil || acc.Username == "" || acc.Password == "" {
		return "", fmt.Errorf("LibreOJ 提交需要配置账号密码")
	}
	lang, ok := lojLangMap[strings.ToLower(req.Language)]
	if !ok {
		return "", fmt.Errorf("LibreOJ 暂不支持语言: %s", req.Language)
	}

	token, err := lojLogin(acc.Username, acc.Password)
	if err != nil {
		return "", err
	}

	displayID, err := strconv.Atoi(strings.TrimSpace(req.RemoteProblemID))
	if err != nil {
		return "", fmt.Errorf("LibreOJ 题号必须是数字: %s", req.RemoteProblemID)
	}
	internalID, err := lojInternalID(displayID)
	if err != nil {
		return "", err
	}

	body, err := lojPost(submitURL, token, map[string]any{
		"problemId": internalID,
		"content": map[string]any{
			"code":                 req.Code,
			"language":             lang.name,
			"compileAndRunOptions": lang.opts,
		},
		"uploadInfo": nil,
	})
	if err != nil {
		return "", err
	}
	var r struct {
		SubmissionID int64 `json:"submissionId"`
	}
	_ = json.Unmarshal(body, &r)
	if r.SubmissionID == 0 {
		return "", fmt.Errorf("LibreOJ 提交失败: %s", string(body))
	}
	return strconv.FormatInt(r.SubmissionID, 10), nil
}

// Poll 查提交详情(公开，无需 token)，把 meta.status 映射成本系统状态码。
func (l *LibreOJ) Poll(_ *remoteoj.RemoteAccount, remoteSubmitID string) (*remoteoj.Verdict, error) {
	body, err := lojPost(detailURL, "", map[string]any{
		"submissionId": remoteSubmitID,
		"locale":       "zh_CN",
	})
	if err != nil {
		return nil, err
	}
	var r struct {
		Meta struct {
			Status     string `json:"status"`
			TimeUsed   int64  `json:"timeUsed"`   // ms
			MemoryUsed int64  `json:"memoryUsed"` // KB
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("解析 LibreOJ 详情失败: %w", err)
	}
	return &remoteoj.Verdict{
		Status:   lojStatus(r.Meta.Status),
		TimeMs:   r.Meta.TimeUsed,
		MemoryKB: r.Meta.MemoryUsed,
		Msg:      r.Meta.Status,
	}, nil
}

func (l *LibreOJ) MapLanguage(lang string) (string, error) {
	if v, ok := lojLangMap[strings.ToLower(lang)]; ok {
		return v.name, nil
	}
	return "", fmt.Errorf("LibreOJ 暂不支持语言: %s", lang)
}

// lojLogin 登录拿 Bearer token
func lojLogin(username, password string) (string, error) {
	body, err := lojPost(loginURL, "", map[string]any{"username": username, "password": password})
	if err != nil {
		return "", err
	}
	var r struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(body, &r)
	if r.Token == "" {
		return "", fmt.Errorf("LibreOJ 登录失败(账号密码错误?): %s", string(body))
	}
	return r.Token, nil
}

// lojInternalID 用 displayId 换内部 problemId(提交接口要的是内部 id)
func lojInternalID(displayID int) (int, error) {
	body, err := lojPost(apiURL, "", map[string]any{"displayId": displayID})
	if err != nil {
		return 0, err
	}
	var r struct {
		Meta struct {
			ID int `json:"id"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(body, &r)
	if r.Meta.ID == 0 {
		return 0, fmt.Errorf("LibreOJ 找不到题目 %d", displayID)
	}
	return r.Meta.ID, nil
}

func lojStatus(s string) int {
	switch s {
	case "Accepted":
		return judge.Accepted
	case "WrongAnswer", "PartiallyCorrect":
		return judge.WrongAnswer
	case "CompilationError":
		return judge.CompileError
	case "TimeLimitExceeded":
		return judge.TimeLimitExceeded
	case "MemoryLimitExceeded":
		return judge.MemoryLimitExceeded
	case "RuntimeError", "OutputLimitExceeded", "FileError":
		return judge.RuntimeError
	case "JudgementFailed", "ConfigurationError", "SystemError", "Canceled":
		return judge.UnknownError
	}
	// Pending / Waiting / Preparing / Compiling / Running ... → 还在判
	return judge.Pending
}

// lojPost 发 JSON POST，token 非空则带 Bearer
func lojPost(url, token string, payload map[string]any) ([]byte, error) {
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
