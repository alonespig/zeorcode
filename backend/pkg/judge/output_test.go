package judge

import "testing"

func TestMD5(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"abc", "900150983cd24fb0d6963f7d28e17f72"},
	}
	for _, tt := range tests {
		if got := MD5(tt.in); got != tt.want {
			t.Errorf("MD5(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRtrimOutput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no trailing whitespace", "1 2\n3", "1 2\n3"},
		{"trailing spaces per line", "1 2  \n3  \n", "1 2\n3"},
		{"trailing newlines", "1 2\n3\n\n\n", "1 2\n3"},
		{"crlf", "1 2\r\n3\r\n", "1 2\n3"},
		{"cr only", "1 2\r3\r", "1 2\n3"},
		{"trailing tabs", "1\t2 \t\n3\t\n", "1\t2\n3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RtrimOutput(tt.in); got != tt.want {
				t.Fatalf("RtrimOutput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStripAllSpace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"spaces", "1 2 3", "123"},
		{"newlines and tabs", "1\n2\t3", "123"},
		{"crlf", "1\r\n2\r\n3", "123"},
		{"no whitespace", "123", "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripAllSpace(tt.in); got != tt.want {
				t.Fatalf("StripAllSpace(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
