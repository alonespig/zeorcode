package judge

import "fmt"

// langSpec 描述一种语言如何在沙箱中编译、运行
type langSpec struct {
	// compile 返回编译用的 cmd payload 以及 copyOutCached 的文件名（即 Artifact.FileID 指向的文件）
	compile func(code string) (cmd map[string]any, cacheName string)
	// run 返回运行命令的 args、要 copyIn 的文件名、proc 上限
	run func(memMB int) (args []string, file string, procLimit int)
}

var languages = map[string]langSpec{
	LanguageCPP: {
		compile: func(code string) (map[string]any, string) {
			return map[string]any{
				// 定义 ONLINE_JUDGE，让用户在本地用的 freopen 等调试块在沙箱里失效
				"args":        []string{"/usr/bin/g++", "a.cc", "-O2", "-std=gnu++17", "-DONLINE_JUDGE", "-o", "a"},
				"env":         []string{"PATH=/usr/bin:/bin", "LANG=C.UTF-8"},
				"files":       []map[string]any{{"content": ""}, {"name": "stdout", "max": 10240}, {"name": "stderr", "max": 1024 * 1024}},
				"cpuLimit":    int64(15_000_000_000),
				"clockLimit":  int64(20_000_000_000),
				"memoryLimit": int64(512 * 1024 * 1024),
				"procLimit":   50,
				"copyIn": map[string]map[string]string{
					"a.cc": {"content": code},
				},
				"copyOut":       []string{"stdout", "stderr"},
				"copyOutCached": []string{"a"},
			}, "a"
		},
		run: func(memMB int) ([]string, string, int) {
			return []string{"a"}, "a", 1
		},
	},

	LanguageJava: {
		compile: func(code string) (map[string]any, string) {
			return map[string]any{
				"args":        []string{"/bin/sh", "-c", "javac Main.java && jar cfe main.jar Main *.class"},
				"env":         []string{"PATH=/usr/bin:/bin", "LANG=C.UTF-8"},
				"files":       []map[string]any{{"content": ""}, {"name": "stdout", "max": 10240}, {"name": "stderr", "max": 2 * 1024 * 1024}},
				"cpuLimit":    int64(20_000_000_000),
				"clockLimit":  int64(30_000_000_000),
				"memoryLimit": int64(1024 * 1024 * 1024),
				"procLimit":   80,
				"copyIn": map[string]map[string]string{
					"Main.java": {"content": code},
				},
				"copyOut":       []string{"stdout", "stderr"},
				"copyOutCached": []string{"main.jar"},
			}, "main.jar"
		},
		run: func(memMB int) ([]string, string, int) {
			heapMB := memMB - 32
			if heapMB < 64 {
				heapMB = 64
			}
			return []string{"java", fmt.Sprintf("-Xmx%dm", heapMB), "-jar", "main.jar"}, "main.jar", 80
		},
	},

	LanguagePython: {
		compile: func(code string) (map[string]any, string) {
			return map[string]any{
				"args":        []string{"python3", "-m", "py_compile", "main.py"},
				"env":         []string{"PATH=/usr/bin:/bin", "LANG=C.UTF-8"},
				"files":       []map[string]any{{"content": ""}, {"name": "stdout", "max": 10240}, {"name": "stderr", "max": 1024 * 1024}},
				"cpuLimit":    int64(5_000_000_000),
				"clockLimit":  int64(8_000_000_000),
				"memoryLimit": int64(256 * 1024 * 1024),
				"procLimit":   20,
				"copyIn": map[string]map[string]string{
					"main.py": {"content": code},
				},
				"copyOut":       []string{"stdout", "stderr"},
				"copyOutCached": []string{"main.py"},
			}, "main.py"
		},
		run: func(memMB int) ([]string, string, int) {
			return []string{"python3", "main.py"}, "main.py", 1
		},
	},
}

// SupportsLanguage 判断名称（含别名）是否有可执行的评测预设。
func SupportsLanguage(language string) bool {
	_, ok := languages[NormalizeLanguage(language)]
	return ok
}

// CanonicalLanguageName 返回数据库和提交记录统一使用的名称。
func CanonicalLanguageName(language string) (string, bool) {
	normalized := NormalizeLanguage(language)
	if _, ok := languages[normalized]; !ok {
		return "", false
	}
	if normalized == LanguageCPP {
		return "c++", true
	}
	return normalized, true
}
