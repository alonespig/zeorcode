// Package publicid generates short numeric identifiers that are safe to expose
// in URLs. They are identifiers, not authorization credentials.
package publicid

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	Min int64 = 10_000_000
	Max int64 = 99_999_999
)

var size = big.NewInt(Max - Min + 1)

// New returns a cryptographically random eight-digit decimal identifier.
func New() (int64, error) {
	n, err := rand.Int(rand.Reader, size)
	if err != nil {
		return 0, fmt.Errorf("generate public id: %w", err)
	}
	return Min + n.Int64(), nil
}

// Valid reports whether id is an eight-digit decimal identifier.
func Valid(id int64) bool {
	return id >= Min && id <= Max
}
