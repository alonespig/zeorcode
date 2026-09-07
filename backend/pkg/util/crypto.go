package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"github.com/spf13/viper"
)

// aesKey 由 config 的 security.aes_key 经 SHA-256 派生为固定 32 字节(AES-256)。
func aesKey() []byte {
	sum := sha256.Sum256([]byte(viper.GetString("security.aes_key")))
	return sum[:]
}

// Encrypt 用 AES-256-GCM 加密明文，返回 base64(nonce|密文|认证标签)。
// 用于落库远程账号的密码 / cookie 等敏感凭证。
func Encrypt(plain string) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt 解密 Encrypt 的输出。
func Decrypt(enc string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("密文长度异常")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func newGCM() (cipher.AEAD, error) {
	block, err := aes.NewCipher(aesKey())
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
