package service

import "golang.org/x/crypto/bcrypt"

// hashPassword 用 bcrypt 生成密码哈希（默认 cost=10）
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword 校验明文密码与 bcrypt 哈希是否匹配
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
