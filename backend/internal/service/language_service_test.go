package service

import (
	"strings"
	"testing"
)

func TestNormalizeLanguageDefaultCode(t *testing.T) {
	t.Run("uses built in template when empty", func(t *testing.T) {
		code, err := normalizeLanguageDefaultCode("c++", "")
		if err != nil {
			t.Fatalf("normalizeLanguageDefaultCode() error = %v", err)
		}
		if !strings.Contains(code, "int main()") {
			t.Fatalf("built-in C++ template missing main: %q", code)
		}
	})

	t.Run("preserves configured template", func(t *testing.T) {
		const want = "custom starter\n"
		got, err := normalizeLanguageDefaultCode("python", want)
		if err != nil {
			t.Fatalf("normalizeLanguageDefaultCode() error = %v", err)
		}
		if got != want {
			t.Fatalf("normalizeLanguageDefaultCode() = %q, want %q", got, want)
		}
	})

	t.Run("rejects oversized template", func(t *testing.T) {
		_, err := normalizeLanguageDefaultCode("java", strings.Repeat("x", maxLanguageDefaultCodeBytes+1))
		if err == nil {
			t.Fatal("normalizeLanguageDefaultCode() expected an error")
		}
	})
}
