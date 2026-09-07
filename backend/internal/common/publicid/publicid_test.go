package publicid

import "testing"

func TestNewReturnsEightDigitID(t *testing.T) {
	for range 100 {
		id, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if !Valid(id) {
			t.Fatalf("New() = %d, want an eight-digit id", id)
		}
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		id   int64
		want bool
	}{
		{Min - 1, false},
		{Min, true},
		{Max, true},
		{Max + 1, false},
	}
	for _, tt := range tests {
		if got := Valid(tt.id); got != tt.want {
			t.Errorf("Valid(%d) = %v, want %v", tt.id, got, tt.want)
		}
	}
}
