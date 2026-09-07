package judge

import "strings"

// Language 支持的语言
const (
	LanguageCPP    = "cpp"
	LanguagePython = "python"
	LanguageJava   = "java"
)

// NormalizeLanguage 统一语言别名 (c++, py, python3...) 到内部规范名
func NormalizeLanguage(language string) string {
	l := strings.ToLower(strings.TrimSpace(language))
	switch l {
	case "c++", "cpp", "cc", "cxx", "g++":
		return LanguageCPP
	case "python", "python3", "py":
		return LanguagePython
	case "java":
		return LanguageJava
	default:
		return l
	}
}

// Limits 单次运行的资源限制
type Limits struct {
	TimeMs   int // CPU 时间上限（毫秒）
	MemoryMB int // 内存上限（MB）
}

// Artifact 编译产物，可在多次 Run 之间复用
type Artifact struct {
	Language string // 规范化后的语言名
	FileID   string // 沙箱缓存的可执行/字节码文件 ID
}

// Verdict 单次运行的判定结果
type Verdict struct {
	Status     int    // judge.Accepted / WrongAnswer / TLE / MLE / RE / UnknownError
	Stdout     string // 标准输出
	Stderr     string // 标准错误
	TimeNs     int64  // 实际运行耗时（纳秒）
	MemoryByte int64  // 实际内存占用（字节）
}

// runResult 与判题机 /run 接口对齐的原始结果（包内私有）
type runResult struct {
	Status     string `json:"status"`
	ExitStatus int    `json:"exitStatus"`
	Time       int64  `json:"time"` // ns, cgroup cpu time
	Memory     int64  `json:"memory"`
	RunTime    int64  `json:"runTime"`
	ProcPeak   int    `json:"procPeak"`
	Files      struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
	} `json:"files"`
	FileIds map[string]string `json:"fileIds"`
}
