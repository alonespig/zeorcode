package nowcoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"

	"zoj/pkg/judge"
	"zoj/pkg/remoteoj"
)

// Nowcoder 牛客 OJ：页面服务端渲染，但站点在阿里云 WAF 后，
// Go 默认 TLS 指纹会被拦，故用系统 curl(指纹被放行)抓 HTML，再 goquery 解析。
// 更干净的做法是用 utls 伪装 Chrome 指纹替掉 curl，见 fetchHTML。
type Nowcoder struct{}

func init() { remoteoj.Register(&Nowcoder{}) }

func (n *Nowcoder) Name() string { return "Nowcoder" }

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var (
	rePid  = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	reTime = regexp.MustCompile(`时间限制：[^，,]*?(\d+)\s*秒`)
	reMem  = regexp.MustCompile(`空间限制：[^，,]*?(\d+)\s*M`)

	// 公式图 <img ... alt="latex" ...> -> $latex$（服务端 HTML 里公式是图片，alt 是 LaTeX 源码）
	reEqImg = regexp.MustCompile(`<img[^>]*\balt="([^"]*)"[^>]*>`)
	reBr    = regexp.MustCompile(`(?i)<br\s*/?>`)
	reTag   = regexp.MustCompile(`<[^>]+>`)

	// 提交要的是页面内部 id（不是 URL 里的 pid），从题目页 HTML/JS 里抓。
	// 形如 "questionId":11714784 / questionId: 11714784 / questionId=11714784 都能匹配。
	reQuestionID = regexp.MustCompile(`questionId["'\s:=]+(\d+)`)
	reTagID      = regexp.MustCompile(`\btagId["'\s:=]+(\d+)`)
)

// ncLang zoj 语言名 -> 牛客 language id + languageName。
// id 取自真实提交的 submit_cd POST 请求体(下拉是 Vue 渲染、HTML 里没有 id)，均已抓包确认。
// 参考：Python2=5、C=3 等如需支持照此扩展。
var ncLang = map[string]struct{ id, name string }{
	"cpp":    {"2", "C++（clang++18）"},
	"java":   {"4", "Java"},
	"python": {"11", "Python3"},
}

func cleanHTML(s string) string {
	s = reEqImg.ReplaceAllString(s, "$$${1}$$") // <img alt> -> $latex$
	s = reBr.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// fetchHTML 用 curl 抓取（绕过阿里云 WAF 对 Go TLS 指纹的拦截）。
// cookie 优先用调用方(service 从 DB 解密后)传入的，为空再回退 config。
func fetchHTML(url, cookie string) ([]byte, error) {
	args := []string{"-sL", "--max-time", "20",
		"-A", browserUA,
		"-H", "Accept-Language: zh-CN,zh;q=0.9",
	}
	if cookie == "" {
		cookie = viper.GetString("remote.nowcoder.cookie")
	}
	// 带浏览器 cookie：稳定过阿里云 WAF + 携带登录态(以后提交复用)
	if ck := strings.TrimSpace(cookie); ck != "" {
		args = append(args, "-b", ck)
	}
	args = append(args, url)

	out, err := exec.Command("curl", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("curl 抓取失败(确认装了 curl): %w", err)
	}
	if bytes.Contains(out, []byte("aliyun_waf")) {
		return nil, fmt.Errorf("被反爬(WAF)拦截，请在 config.yaml 配置 remote.nowcoder.cookie")
	}
	return out, nil
}

// curlPostForm 用 curl 发 x-www-form-urlencoded POST。
// 牛客提交接口(submit_cd)是 POST + AJAX，参数在请求体里。
func curlPostForm(endpoint string, form url.Values, cookie, referer string) ([]byte, error) {
	args := []string{"-sL", "--max-time", "20",
		"-A", browserUA,
		"-H", "Accept-Language: zh-CN,zh;q=0.9",
		"-H", "X-Requested-With: XMLHttpRequest",
		"-H", "Content-Type: application/x-www-form-urlencoded; charset=UTF-8",
	}
	if referer != "" {
		args = append(args, "-H", "Referer: "+referer)
	}
	if cookie == "" {
		cookie = viper.GetString("remote.nowcoder.cookie")
	}
	if ck := strings.TrimSpace(cookie); ck != "" {
		args = append(args, "-b", ck)
	}
	args = append(args, "--data", form.Encode(), endpoint)

	out, err := exec.Command("curl", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("curl 提交失败: %w", err)
	}
	if bytes.Contains(out, []byte("aliyun_waf")) {
		return nil, fmt.Errorf("被反爬(WAF)拦截，请检查 cookie")
	}
	return out, nil
}

func (n *Nowcoder) CrawlProblem(acc *remoteoj.RemoteAccount, pid string) (*remoteoj.RemoteProblem, error) {
	if !rePid.MatchString(pid) {
		return nil, fmt.Errorf("非法题号: %q", pid)
	}
	cookie := ""
	if acc != nil {
		cookie = acc.Cookie
	}
	url := fmt.Sprintf("https://ac.nowcoder.com/acm/problem/%s", pid)
	htmlBytes, err := fetchHTML(url, cookie)
	if err != nil {
		return nil, err
	}
	return parseProblem(htmlBytes, pid, url)
}

// parseProblem 纯解析（与抓取解耦，便于离线单测）
func parseProblem(htmlBytes []byte, pid, url string) (*remoteoj.RemoteProblem, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find(".question-title").Text())

	info := doc.Text()
	timeLimit, memoryLimit := 1000, 256
	if m := reTime.FindStringSubmatch(info); m != nil {
		if s, _ := strconv.Atoi(m[1]); s > 0 {
			timeLimit = s * 1000
		}
	}
	if m := reMem.FindStringSubmatch(info); m != nil {
		if s, _ := strconv.Atoi(m[1]); s > 0 {
			memoryLimit = s
		}
	}

	descHTML, _ := doc.Find(".subject-question").Html()
	description := cleanHTML(descHTML)

	var inputFormat, outputFormat string
	doc.Find(".subject-describe h2").Each(func(_ int, h *goquery.Selection) {
		t := strings.TrimSpace(h.Text())
		switch {
		case strings.HasPrefix(t, "输入描述"):
			ih, _ := h.Next().Html()
			inputFormat = cleanHTML(ih)
		case strings.HasPrefix(t, "输出描述"):
			oh, _ := h.Next().Html()
			outputFormat = cleanHTML(oh)
		}
	})

	var samples []remoteoj.ProblemSample
	for i := 1; ; i++ {
		in := doc.Find(fmt.Sprintf(`textarea[data-clipboard-text-id="input%d"]`, i))
		if in.Length() == 0 {
			break
		}
		out := doc.Find(fmt.Sprintf(`textarea[data-clipboard-text-id="output%d"]`, i))
		samples = append(samples, remoteoj.ProblemSample{
			Input:  strings.TrimRight(in.Text(), "\n"),
			Output: strings.TrimRight(out.Text(), "\n"),
		})
	}

	if title == "" {
		return nil, fmt.Errorf("解析失败，可能题目不存在或页面结构变化 (pid=%s)", pid)
	}

	return &remoteoj.RemoteProblem{
		OJ:              "Nowcoder",
		RemoteProblemID: pid,
		Title:           title,
		TimeLimit:       timeLimit,
		MemoryLimit:     memoryLimit,
		Description:     description,
		InputFormat:     inputFormat,
		OutputFormat:    outputFormat,
		Samples:         samples,
		Source:          url,
	}, nil
}

// Submit 提交代码到牛客，返回远程提交号(内部编码了 tagId，Poll 时复用)。
func (n *Nowcoder) Submit(acc *remoteoj.RemoteAccount, req remoteoj.SubmitReq) (string, error) {
	if acc == nil || strings.TrimSpace(acc.Cookie) == "" {
		return "", fmt.Errorf("牛客提交需要登录 cookie")
	}
	lang, ok := ncLang[strings.ToLower(req.Language)]
	if !ok {
		return "", fmt.Errorf("牛客暂不支持语言: %s", req.Language)
	}

	// 1. 从题目页抓 questionId / tagId（提交接口要的内部 id）
	page, err := fetchHTML(fmt.Sprintf("https://ac.nowcoder.com/acm/problem/%s", req.RemoteProblemID), acc.Cookie)
	if err != nil {
		return "", err
	}
	qid := reQuestionID.FindSubmatch(page)
	tid := reTagID.FindSubmatch(page)
	if qid == nil || tid == nil {
		return "", fmt.Errorf("解析 questionId/tagId 失败，页面结构可能变化")
	}
	questionID, tagID := string(qid[1]), string(tid[1])

	// 2. 提交（POST，参数在请求体；与浏览器抓包一致）
	q := url.Values{}
	q.Set("questionId", questionID)
	q.Set("tagId", tagID)
	q.Set("subTagId", "0")
	q.Set("content", req.Code)
	q.Set("language", lang.id)
	q.Set("languageName", lang.name)
	q.Set("doneQuestionId", req.RemoteProblemID)

	referer := fmt.Sprintf("https://ac.nowcoder.com/acm/problem/%s", req.RemoteProblemID)
	body, err := curlPostForm("https://ac.nowcoder.com/nccommon/submit_cd", q, acc.Cookie, referer)
	if err != nil {
		return "", err
	}
	var sr struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data int64  `json:"data"` // submissionId
	}
	if err := json.Unmarshal(body, &sr); err != nil {
		return "", fmt.Errorf("提交响应解析失败: %w", err)
	}
	if sr.Code != 0 || sr.Data == 0 {
		return "", fmt.Errorf("牛客提交失败: %s", sr.Msg)
	}
	return fmt.Sprintf("%d|%s", sr.Data, tagID), nil
}

// Poll 查询远程评测结果并映射为本系统状态码。Status==Pending 表示还在判，调用方继续轮询。
func (n *Nowcoder) Poll(acc *remoteoj.RemoteAccount, remoteSubmitID string) (*remoteoj.Verdict, error) {
	if acc == nil || strings.TrimSpace(acc.Cookie) == "" {
		return nil, fmt.Errorf("牛客查询需要登录 cookie")
	}
	subID, tagID := remoteSubmitID, "0"
	if i := strings.IndexByte(remoteSubmitID, '|'); i >= 0 {
		subID, tagID = remoteSubmitID[:i], remoteSubmitID[i+1:]
	}

	q := url.Values{}
	q.Set("submissionId", subID)
	q.Set("tagId", tagID)
	q.Set("subTagId", "0")

	body, err := fetchHTML("https://ac.nowcoder.com/nccommon/status?"+q.Encode(), acc.Cookie)
	if err != nil {
		return nil, err
	}
	var sr struct {
		Code              int    `json:"code"`
		Status            int    `json:"status"`
		JudgeReplyDesc    string `json:"judgeReplyDesc"`
		Desc              string `json:"desc"`
		Memo              string `json:"memo"`
		TimeConsumption   int64  `json:"timeConsumption"`   // ms
		MemoryConsumption int64  `json:"memoryConsumption"` // KB
	}
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("查询响应解析失败: %w", err)
	}

	return &remoteoj.Verdict{
		Status:   mapStatus(sr.Status, sr.JudgeReplyDesc),
		TimeMs:   sr.TimeConsumption,
		MemoryKB: sr.MemoryConsumption,
		Msg:      firstNonEmpty(sr.Memo, sr.Desc, sr.JudgeReplyDesc),
	}, nil
}

func (n *Nowcoder) MapLanguage(lang string) (string, error) {
	if l, ok := ncLang[strings.ToLower(lang)]; ok {
		return l.id, nil
	}
	return "", fmt.Errorf("牛客暂不支持语言: %s", lang)
}

// mapStatus 牛客结果 -> zoj 状态码。优先按中文判定描述匹配(更稳)，再用已知状态码兜底。
func mapStatus(status int, desc string) int {
	switch {
	case strings.Contains(desc, "答案正确"), strings.Contains(desc, "通过"):
		return judge.Accepted
	case strings.Contains(desc, "编译错误"):
		return judge.CompileError
	case strings.Contains(desc, "答案错误"), strings.Contains(desc, "部分正确"), strings.Contains(desc, "格式错误"):
		return judge.WrongAnswer
	case strings.Contains(desc, "超时"), strings.Contains(desc, "时间超限"):
		return judge.TimeLimitExceeded
	case strings.Contains(desc, "内存"), strings.Contains(desc, "空间超限"):
		return judge.MemoryLimitExceeded
	case strings.Contains(desc, "运行错误"), strings.Contains(desc, "段错误"), strings.Contains(desc, "异常"):
		return judge.RuntimeError
	case strings.Contains(desc, "等待"), strings.Contains(desc, "排队"),
		strings.Contains(desc, "判题中"), strings.Contains(desc, "运行中"), strings.Contains(desc, "正在"):
		return judge.Pending
	}
	switch status {
	case 5:
		return judge.Accepted
	case 12:
		return judge.CompileError // 已确认
	case 0, 1, 2, 3, 4:
		return judge.Pending
	}
	return judge.UnknownError
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
