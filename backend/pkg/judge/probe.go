package judge

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ProbeResult 一台评测机的探活结果（后台健康面板用）
type ProbeResult struct {
	URL       string `json:"url"`
	Online    bool   `json:"online"`
	LatencyMs int64  `json:"latencyMs"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Probe 现场探一台 go-judge：能拿到 HTTP 响应即视为在线，尽量读出版本与延迟。
func Probe(rawURL string) ProbeResult {
	base := strings.TrimRight(strings.TrimSpace(rawURL), "/")
	res := ProbeResult{URL: base}
	hc := &http.Client{Timeout: 5 * time.Second}

	start := time.Now()
	resp, err := hc.Get(base + "/version")
	if err != nil {
		// /version 不通就退而探根路径，只判可达
		resp2, err2 := hc.Get(base)
		res.LatencyMs = time.Since(start).Milliseconds()
		if err2 != nil {
			res.Error = err2.Error()
			return res
		}
		_ = resp2.Body.Close()
		res.Online = true
		return res
	}
	res.LatencyMs = time.Since(start).Milliseconds()
	defer resp.Body.Close()
	res.Online = true

	// best-effort 解析 go-judge /version 里的 buildVersion
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var v map[string]any
	if json.Unmarshal(body, &v) == nil {
		if bv, ok := v["buildVersion"].(string); ok {
			res.Version = bv
		}
	}
	return res
}

// ProbeAll 并发探活一组地址，结果按入参顺序返回。
func ProbeAll(urls []string) []ProbeResult {
	results := make([]ProbeResult, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			results[i] = Probe(u)
		}(i, u)
	}
	wg.Wait()
	return results
}
