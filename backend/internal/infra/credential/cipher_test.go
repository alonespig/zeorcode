package credential

import (
	"encoding/base64"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plain := "cookie-token-123"
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if dec != plain {
		t.Fatalf("roundtrip = %q, want %q", dec, plain)
	}
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	a, err := Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	b, err := Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if a == b {
		t.Fatal("Encrypt() produced identical ciphertext for two calls, nonce is not random")
	}
}

func TestDecryptRejectsInvalidBase64(t *testing.T) {
	if _, err := Decrypt("not-valid-base64!!!"); err == nil {
		t.Fatal("Decrypt() error = nil, want invalid base64 error")
	}
}

func TestDecryptRejectsTooShortCiphertext(t *testing.T) {
	// 3 字节明文 < GCM nonce 大小(12)，应被拒绝
	short := base64.StdEncoding.EncodeToString([]byte("abc"))
	if _, err := Decrypt(short); err == nil {
		t.Fatal("Decrypt() error = nil, want short ciphertext error")
	}
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	enc, err := Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatalf("base64 decode error = %v", err)
	}
	data[len(data)-1] ^= 0xff // 破坏密文/认证标签
	tampered := base64.StdEncoding.EncodeToString(data)
	if _, err := Decrypt(tampered); err == nil {
		t.Fatal("Decrypt() error = nil, want authentication failure")
	}
}
