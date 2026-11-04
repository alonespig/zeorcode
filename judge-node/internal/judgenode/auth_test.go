package judgenode

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestAuthorized(t *testing.T) {
	t.Parallel()

	const token = "judge-node-secret"
	tests := []struct {
		name     string
		values   []string
		expected string
		want     bool
	}{
		{name: "valid", values: []string{"Bearer " + token}, expected: token, want: true},
		{name: "case insensitive scheme", values: []string{"bearer " + token}, expected: token, want: true},
		{name: "wrong token", values: []string{"Bearer wrong"}, expected: token},
		{name: "missing bearer", values: []string{token}, expected: token},
		{name: "duplicate header", values: []string{"Bearer " + token, "Bearer " + token}, expected: token},
		{name: "empty server token", values: []string{"Bearer " + token}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"authorization": test.values})
			if got := authorized(ctx, test.expected); got != test.want {
				t.Fatalf("authorized() = %v, want %v", got, test.want)
			}
		})
	}
}
