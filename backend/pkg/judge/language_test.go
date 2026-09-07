package judge

import "testing"

func TestCanonicalLanguageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: " C++ ", want: "c++", ok: true},
		{input: "cpp", want: "c++", ok: true},
		{input: "Python3", want: "python", ok: true},
		{input: "java", want: "java", ok: true},
		{input: "go", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := CanonicalLanguageName(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("CanonicalLanguageName(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}
