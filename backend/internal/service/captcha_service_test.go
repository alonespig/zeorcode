package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"zoj/internal/common/errcode"

	goredis "github.com/redis/go-redis/v9"
)

type fakeCaptchaStore struct {
	values map[string]string
	ttl    time.Duration
}

func (f *fakeCaptchaStore) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	if f.values == nil {
		f.values = make(map[string]string)
	}
	f.values[key] = value.(string)
	f.ttl = ttl
	return nil
}

func (f *fakeCaptchaStore) GetAndDelete(_ context.Context, key string) (string, error) {
	value, ok := f.values[key]
	if !ok {
		return "", goredis.Nil
	}
	delete(f.values, key)
	return value, nil
}

func TestCaptchaGenerateAndVerifyOnce(t *testing.T) {
	store := &fakeCaptchaStore{}
	service := &CaptchaService{store: store}

	resp, err := service.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if resp.ID == "" || !strings.HasPrefix(resp.Image, "data:image/png;base64,") {
		t.Fatalf("Generate() response = %#v", resp)
	}
	if store.ttl != captchaTTL {
		t.Fatalf("captcha TTL = %s, want %s", store.ttl, captchaTTL)
	}
	answer := store.values[captchaKey(resp.ID)]
	if err := service.Verify(context.Background(), resp.ID, answer); err != nil {
		t.Fatalf("Verify() first error = %v", err)
	}
	if err := service.Verify(context.Background(), resp.ID, answer); !errors.Is(err, errcode.ErrCaptchaInvalid) {
		t.Fatalf("Verify() second error = %v, want captcha invalid", err)
	}
}

func TestCaptchaVerifyConsumesWrongAnswer(t *testing.T) {
	store := &fakeCaptchaStore{values: map[string]string{captchaKey("id"): "12345"}}
	service := &CaptchaService{store: store}

	if err := service.Verify(context.Background(), "id", "00000"); !errors.Is(err, errcode.ErrCaptchaInvalid) {
		t.Fatalf("Verify() error = %v, want captcha invalid", err)
	}
	if _, ok := store.values[captchaKey("id")]; ok {
		t.Fatal("wrong answer did not consume captcha")
	}
}
