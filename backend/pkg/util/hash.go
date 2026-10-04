package util

import (
	"crypto/md5"
	"fmt"
)

// MD5 取字符串的 MD5 十六进制摘要
func MD5(str string) string {
	sum := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", sum)
}
