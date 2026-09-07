package atcoder

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"zoj/pkg/remoteoj"
)

// AtCoder 实现。页面是普通服务端渲染(无阿里云式 JS 挑战)，公开题直接 GET 即可。
// 提交表单挂了 Cloudflare Turnstile 验证码，无法自动提交，故 Submit 不实现。
type AtCoder struct{}

func init() { remoteoj.Register(&AtCoder{}) }

func (a *AtCoder) Name() string { return "AtCoder" }

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var (
	reTime       = regexp.MustCompile(`Time Limit:\s*([\d.]+)\s*sec`)
	reMem        = regexp.MustCompile(`Memory Limit:\s*(\d+)\s*MiB`)
	reTailNum    = regexp.MustCompile(`(\d+)\s*$`)
	reBr         = regexp.MustCompile(`(?i)<br\s*/?>`)
	reTag        = regexp.MustCompile(`<[^>]+>`)
	reBlankLines = regexp.MustCompile(`\n{3,}`)
)

// acLang zoj 语言名 -> AtCoder LanguageId(取自提交页 <select> 的 option value)
var acLang = map[string]string{
	"cpp":    "6017", // C++23 (GCC 15.2.0)
	"java":   "6056", // Java24 (OpenJDK 24.0.2)
	"python": "6082", // Python (CPython 3.13.7)
}

func (a *AtCoder) CrawlProblem(acc *remoteoj.RemoteAccount, pid string) (*remoteoj.RemoteProblem, error) {
	contest, task := splitPID(pid)
	if contest == "" || task == "" {
		return nil, fmt.Errorf("非法题号: %q（应形如 abc462_a 或 abc462/abc462_a）", pid)
	}
	url := fmt.Sprintf("https://atcoder.jp/contests/%s/tasks/%s?lang=en", contest, task)
	body, err := fetchHTML(url, cookieOf(acc))
	if err != nil {
		return nil, err
	}
	return parseProblem(body, task, url)
}

// splitPID 解析题号：支持 "abc462/abc462_a" 或 "abc462_a"(由 task 反推 contest)。
func splitPID(pid string) (contest, task string) {
	pid = strings.TrimSpace(pid)
	if i := strings.Index(pid, "/"); i >= 0 {
		return pid[:i], pid[i+1:]
	}
	if i := strings.LastIndex(pid, "_"); i > 0 {
		return pid[:i], pid
	}
	return "", pid
}

func cookieOf(acc *remoteoj.RemoteAccount) string {
	if acc != nil {
		return acc.Cookie
	}
	return ""
}

// fetchHTML 用 Go 原生 http 抓页面(AtCoder 不做 TLS 指纹拦截)。
// cookie 可选：比赛进行中的题或私有题需要登录态(REVEL_SESSION)。
func fetchHTML(url, cookie string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept-Language", "en")
	if c := strings.TrimSpace(cookie); c != "" {
		req.Header.Set("Cookie", c)
	}

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	case http.StatusNotFound:
		return nil, fmt.Errorf("题目不存在(404)")
	default:
		return nil, fmt.Errorf("抓取失败, http %d (比赛进行中的题需要配置登录 cookie)", resp.StatusCode)
	}
}

// parseProblem 解析英文题面(与抓取解耦，便于离线单测)。
func parseProblem(htmlBytes []byte, task, url string) (*remoteoj.RemoteProblem, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", ""))

	page := doc.Text()
	timeLimit, memLimit := 2000, 256
	if m := reTime.FindStringSubmatch(page); m != nil {
		if sec, e := strconv.ParseFloat(m[1], 64); e == nil && sec > 0 {
			timeLimit = int(sec * 1000)
		}
	}
	if m := reMem.FindStringSubmatch(page); m != nil {
		if mb, e := strconv.Atoi(m[1]); e == nil && mb > 0 {
			memLimit = mb
		}
	}

	// 题面有日/英两份，取英文；没有就退回整块
	root := doc.Find("#task-statement span.lang-en")
	if root.Length() == 0 {
		root = doc.Find("#task-statement")
	}

	var description, constraints, inputFmt, outputFmt string
	inputs, outputs := map[int]string{}, map[int]string{}

	root.Find("div.part > section").Each(func(_ int, sec *goquery.Selection) {
		h := strings.TrimSpace(sec.Find("h3").First().Text())
		switch {
		case h == "Problem Statement":
			description = sectionText(sec)
		case h == "Constraints":
			constraints = sectionText(sec)
		case h == "Input":
			inputFmt = sectionText(sec)
		case h == "Output":
			outputFmt = sectionText(sec)
		case strings.HasPrefix(h, "Sample Input"):
			inputs[tailNum(h)] = preText(sec)
		case strings.HasPrefix(h, "Sample Output"):
			outputs[tailNum(h)] = preText(sec)
		}
	})

	if constraints != "" {
		description = strings.TrimSpace(description + "\n\n## Constraints\n" + constraints)
	}

	var samples []remoteoj.ProblemSample
	for i := 1; ; i++ {
		in, ok := inputs[i]
		if !ok {
			break
		}
		samples = append(samples, remoteoj.ProblemSample{Input: in, Output: outputs[i]})
	}

	if title == "" {
		return nil, fmt.Errorf("解析失败，可能题目不存在或页面结构变化 (task=%s)", task)
	}

	return &remoteoj.RemoteProblem{
		OJ:              "AtCoder",
		RemoteProblemID: task,
		Title:           title,
		TimeLimit:       timeLimit,
		MemoryLimit:     memLimit,
		Description:     description,
		InputFormat:     inputFmt,
		OutputFormat:    outputFmt,
		Samples:         samples,
		Source:          url,
	}, nil
}

func tailNum(s string) int {
	if m := reTailNum.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

func preText(sec *goquery.Selection) string {
	return strings.TrimRight(sec.Find("pre").First().Text(), "\n")
}

// sectionText 取 section 去掉 h3 后的纯文本
func sectionText(sec *goquery.Selection) string {
	clone := sec.Clone()
	clone.Find("h3").Remove()
	h, _ := clone.Html()
	return cleanHTML(h)
}

func cleanHTML(s string) string {
	s = reBr.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = reBlankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// Submit AtCoder 提交受 Cloudflare Turnstile 验证码保护，无法自动提交。
func (a *AtCoder) Submit(acc *remoteoj.RemoteAccount, req remoteoj.SubmitReq) (string, error) {
	return "", fmt.Errorf("AtCoder 提交受 Cloudflare Turnstile 验证码保护，暂不支持自动提交")
}

func (a *AtCoder) Poll(acc *remoteoj.RemoteAccount, remoteSubmitID string) (*remoteoj.Verdict, error) {
	return nil, remoteoj.ErrNotImplemented
}

func (a *AtCoder) MapLanguage(lang string) (string, error) {
	if id, ok := acLang[strings.ToLower(lang)]; ok {
		return id, nil
	}
	return "", fmt.Errorf("AtCoder 暂不支持语言: %s", lang)
}
