package hdu

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"

	"zoj/pkg/judge"
	"zoj/pkg/remoteoj"
)

const hduHost = "https://acm.hdu.edu.cn"

// HDU 杭电 OJ。页面是老式服务端渲染、编码为 GB2312(需转码)，公开题直接 GET。
// 题面用 $...$(MathJax) + **markdown**，去标签后原样保留，正好对接 zoj 的 markdown+katex 渲染。
// 提交需登录且有验证码，不做远程判题 —— HDU 定位为"导题源"，导入后本地评测。
type HDU struct{}

func init() { remoteoj.Register(&HDU{}) }

func (h *HDU) Name() string { return "HDU" }

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var (
	rePid = regexp.MustCompile(`^\d+$`)
	// "6000/3000 MS" 取 Others(斜杠后)；"1000 MS" 取该值
	reTime  = regexp.MustCompile(`Time Limit:\s*(?:\d+/)?(\d+)\s*MS`)
	reMem   = regexp.MustCompile(`Memory Limit:\s*(?:\d+/)?(\d+)\s*K`)
	reBr    = regexp.MustCompile(`(?i)<br\s*/?>`)
	reTag   = regexp.MustCompile(`<[^>]+>`)
	reBlank = regexp.MustCompile(`\n{3,}`)

	// 远程判题用：从 status.php 抓 run id / 结果行；viewerror.php 抓 CE 信息
	reRunID = regexp.MustCompile(`<td height=22px>(\d+)`)
	rePre   = regexp.MustCompile(`<pre>([\s\S]*?)</pre>`)
)

func (h *HDU) CrawlProblem(_ *remoteoj.RemoteAccount, pid string) (*remoteoj.RemoteProblem, error) {
	if !rePid.MatchString(pid) {
		return nil, fmt.Errorf("HDU 题号必须是数字: %q", pid)
	}
	url := fmt.Sprintf("https://acm.hdu.edu.cn/showproblem.php?pid=%s", pid)
	body, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}
	return parseProblem(body, pid, url)
}

// fetchHTML 抓取并把 GB2312 转成 UTF-8(charset.NewReader 按 meta/Content-Type 自动转码)。
func fetchHTML(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("抓取失败, http %d", resp.StatusCode)
	}

	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("编码转换失败: %w", err)
	}
	return io.ReadAll(utf8Reader)
}

// parseProblem 解析 HDU 题面(与抓取解耦，便于离线单测)。
func parseProblem(htmlBytes []byte, pid, url string) (*remoteoj.RemoteProblem, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find("h1").First().Text())

	page := doc.Text()
	timeLimit, memLimit := 1000, 256
	if m := reTime.FindStringSubmatch(page); m != nil {
		if ms, e := strconv.Atoi(m[1]); e == nil && ms > 0 {
			timeLimit = ms
		}
	}
	if m := reMem.FindStringSubmatch(page); m != nil {
		if k, e := strconv.Atoi(m[1]); e == nil && k > 0 {
			memLimit = k / 1024 // K -> MB
		}
	}

	var description, inputFmt, outputFmt, hint, sampleIn, sampleOut string
	// 每个 panel_title 紧跟一个 panel_content
	doc.Find("div.panel_title").Each(func(_ int, t *goquery.Selection) {
		head := strings.TrimSpace(t.Text())
		content := t.Next()
		if !content.HasClass("panel_content") {
			return
		}
		switch head {
		case "Problem Description":
			description = sectionText(content)
		case "Input":
			inputFmt = sectionText(content)
		case "Output":
			outputFmt = sectionText(content)
		case "Hint":
			hint = sectionText(content)
		case "Sample Input":
			sampleIn = preText(content)
		case "Sample Output":
			sampleOut = preText(content)
		}
	})

	var samples []remoteoj.ProblemSample
	if sampleIn != "" || sampleOut != "" {
		samples = append(samples, remoteoj.ProblemSample{Input: sampleIn, Output: sampleOut})
	}

	if title == "" {
		return nil, fmt.Errorf("解析失败，可能题目不存在或页面结构变化 (pid=%s)", pid)
	}

	return &remoteoj.RemoteProblem{
		OJ:              "HDU",
		RemoteProblemID: pid,
		Title:           title,
		TimeLimit:       timeLimit,
		MemoryLimit:     memLimit,
		Description:     description,
		InputFormat:     inputFmt,
		OutputFormat:    outputFmt,
		Samples:         samples,
		Hint:            hint,
		Source:          url,
	}, nil
}

// preText 取样例 <pre> 的原始文本(保留换行，去首尾空行)
func preText(content *goquery.Selection) string {
	return strings.Trim(content.Find("pre").First().Text(), "\n")
}

// sectionText 取题面文本：<br> 转换行、去标签，保留 $LaTeX$ 与 **markdown**
func sectionText(content *goquery.Selection) string {
	h, err := content.Html()
	if err != nil {
		return strings.TrimSpace(content.Text())
	}
	h = reBr.ReplaceAllString(h, "\n")
	h = reTag.ReplaceAllString(h, "")
	h = html.UnescapeString(h)
	h = reBlank.ReplaceAllString(h, "\n\n")
	return strings.TrimSpace(h)
}

// Submit 账号密码登录 → 提交代码 → 反查最新 run id 当远程提交号。
// HDU 登录只要账号密码(无验证码)，提交后只给 302，提交号靠查"该账号该题最新一条"反推。
func (h *HDU) Submit(acc *remoteoj.RemoteAccount, req remoteoj.SubmitReq) (string, error) {
	if acc == nil || acc.Username == "" || acc.Password == "" {
		return "", fmt.Errorf("HDU 提交需要配置账号密码")
	}
	langID, err := h.MapLanguage(req.Language)
	if err != nil {
		return "", err
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		// 不自动跟随跳转：登录/提交成功都是 302，需要拿到 302 并让 cookie 落到 jar
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	if err := hduLogin(client, acc.Username, acc.Password); err != nil {
		return "", err
	}

	// 代码尾部加随机空白 → 每次源码唯一，绕过"请勿重复提交相同代码"
	form := url.Values{}
	form.Set("check", "0")
	form.Set("language", langID)
	form.Set("problemid", req.RemoteProblemID)
	form.Set("_usercode", base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(req.Code+randomBlank()))))

	submitURL := hduHost + "/submit.php?action=submit"
	status, body, err := hduPostForm(client, submitURL, form)
	if err != nil {
		return "", err
	}
	if status == http.StatusOK && strings.Contains(body, "Please don't re-submit") {
		time.Sleep(5 * time.Second) // 提交频率限制，等一下重试
		status, _, err = hduPostForm(client, submitURL, form)
		if err != nil {
			return "", err
		}
	}
	if status != http.StatusFound {
		return "", fmt.Errorf("HDU 提交失败, http %d", status)
	}

	runID := hduMaxRunID(acc.Username, req.RemoteProblemID)
	if runID == "" {
		time.Sleep(2 * time.Second)
		runID = hduMaxRunID(acc.Username, req.RemoteProblemID)
	}
	if runID == "" {
		return "", fmt.Errorf("HDU 提交后未取到 run id")
	}
	return runID, nil
}

// Poll 查 status.php(公开页，无需登录)，正则抠该 run id 的状态/时间/内存。
func (h *HDU) Poll(_ *remoteoj.RemoteAccount, remoteSubmitID string) (*remoteoj.Verdict, error) {
	body, err := fetchHTML(fmt.Sprintf("%s/status.php?first=%s", hduHost, remoteSubmitID))
	if err != nil {
		return nil, err
	}

	row := regexp.MustCompile(`>` + regexp.QuoteMeta(remoteSubmitID) +
		`</td><td>[\s\S]*?</td><td>([\s\S]*?)</td><td>[\s\S]*?</td><td>(\d*?)MS</td><td>(\d*?)K</td>`)
	m := row.FindSubmatch(body)
	if m == nil {
		// 还没出现在列表里 → 当作还在判，让轮询继续
		return &remoteoj.Verdict{Status: judge.Pending}, nil
	}

	rawStatus := strings.TrimSpace(reTag.ReplaceAllString(string(m[1]), ""))
	v := &remoteoj.Verdict{Status: hduStatus(rawStatus), Msg: rawStatus}
	if v.Status == judge.Pending {
		return v, nil
	}
	if t, e := strconv.Atoi(string(m[2])); e == nil {
		v.TimeMs = int64(t)
	}
	if mem, e := strconv.Atoi(string(m[3])); e == nil {
		v.MemoryKB = int64(mem)
	}
	if v.Status == judge.CompileError {
		if ce, e := fetchHTML(fmt.Sprintf("%s/viewerror.php?rid=%s", hduHost, remoteSubmitID)); e == nil {
			if pm := rePre.FindSubmatch(ce); pm != nil {
				v.Msg = html.UnescapeString(string(pm[1]))
			}
		}
	}
	return v, nil
}

func (h *HDU) MapLanguage(lang string) (string, error) {
	switch strings.ToLower(lang) {
	case "cpp":
		return "0", nil // G++
	case "c":
		return "3", nil
	case "java":
		return "5", nil
	default:
		return "", fmt.Errorf("HDU 暂不支持语言: %s", lang)
	}
}

func hduLogin(client *http.Client, user, pass string) error {
	form := url.Values{}
	form.Set("username", user)
	form.Set("userpass", pass)
	form.Set("login", "Sign In")
	status, _, err := hduPostForm(client, hduHost+"/userloginex.php?action=login", form)
	if err != nil {
		return err
	}
	if status != http.StatusFound { // 成功是 302
		return fmt.Errorf("HDU 登录失败(账号或密码错误?), http %d", status)
	}
	return nil
}

// hduPostForm POST 表单，返回状态码 + GB2312 解码后的 body
func hduPostForm(client *http.Client, target string, form url.Values) (int, string, error) {
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Referer", hduHost)
	req.Header.Set("Origin", hduHost)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return resp.StatusCode, "", nil
	}
	b, _ := io.ReadAll(r)
	return resp.StatusCode, string(b), nil
}

// hduMaxRunID 查该账号该题最新一条提交的 run id(公开页)
func hduMaxRunID(user, pid string) string {
	body, err := fetchHTML(fmt.Sprintf("%s/status.php?user=%s&pid=%s", hduHost, url.QueryEscape(user), pid))
	if err != nil {
		return ""
	}
	if m := reRunID.FindSubmatch(body); m != nil {
		return string(m[1])
	}
	return ""
}

func hduStatus(raw string) int {
	if strings.Contains(raw, "Runtime Error") {
		return judge.RuntimeError
	}
	switch raw {
	case "Accepted":
		return judge.Accepted
	case "Wrong Answer":
		return judge.WrongAnswer
	case "Compilation Error":
		return judge.CompileError
	case "Time Limit Exceeded":
		return judge.TimeLimitExceeded
	case "Memory Limit Exceeded":
		return judge.MemoryLimitExceeded
	case "Output Limit Exceeded":
		return judge.RuntimeError
	case "Presentation Error": // zoj 无 PE，归到 WA
		return judge.WrongAnswer
	}
	// Queuing / Running / Compiling / Submitted 等 → 还在判
	return judge.Pending
}

// randomBlank 追加随机空白(空格/制表)，让每次提交源码唯一
func randomBlank() string {
	n := time.Now().UnixNano()
	var b strings.Builder
	b.WriteByte('\n')
	for n > 0 {
		if n%2 == 0 {
			b.WriteByte(' ')
		} else {
			b.WriteByte('\t')
		}
		n /= 2
	}
	return b.String()
}
