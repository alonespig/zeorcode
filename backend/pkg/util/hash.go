package util

import (
	"crypto/md5"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 用 bcrypt 生成密码哈希（默认 cost=10）
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验明文密码与 bcrypt 哈希是否匹配
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5 取字符串的 MD5 十六进制摘要
func MD5(str string) string {
	sum := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", sum)
}
