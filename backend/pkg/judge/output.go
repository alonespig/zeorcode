package judge

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// MD5 取字符串的 MD5 十六进制摘要
func MD5(str string) string {
	sum := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", sum)
}

// RtrimOutput 去掉每行行末空白 + 文末空白（容忍行末空格、末尾多余换行），
// 用于判题输出与期望输出的 AC 比对。
func RtrimOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t\f\v")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// StripAllSpace 去掉所有空白字符（用于 PE 判定：内容一致仅空白不同）
func StripAllSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}
