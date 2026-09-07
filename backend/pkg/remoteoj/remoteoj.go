package remoteoj

import (
	"errors"
	"sort"
	"strings"
)

// ErrNotImplemented 占位实现可返回此错误
var ErrNotImplemented = errors.New("remoteoj: not implemented")

// RemoteOJ 远程 OJ 策略接口：每个外部 OJ 实现一份。
// CrawlProblem 的 acc 由 service 层从 DB 取出并解密 cookie 后传入，
// 公开题可传 nil；判题(Submit/Poll)必须带账号(cookie)。
type RemoteOJ interface {
	Name() string
	CrawlProblem(acc *RemoteAccount, pid string) (*RemoteProblem, error)
	Submit(acc *RemoteAccount, req SubmitReq) (remoteSubmitID string, err error)
	Poll(acc *RemoteAccount, remoteSubmitID string) (*Verdict, error)
	MapLanguage(lang string) (string, error)
}

var registry = map[string]RemoteOJ{}

// Register 注册一个 OJ 实现（各实现在 init() 里调用）
func Register(oj RemoteOJ) {
	registry[strings.ToUpper(oj.Name())] = oj
}

// Get 按名取实现
func Get(name string) (RemoteOJ, bool) {
	oj, ok := registry[strings.ToUpper(name)]
	return oj, ok
}

// Supported 返回所有已注册 OJ 名（给前端下拉用）
func Supported() []string {
	names := make([]string, 0, len(registry))
	for _, oj := range registry {
		names = append(names, oj.Name())
	}
	sort.Strings(names)
	return names
}
