// Package ids provides type-safe UUID-based identifiers.
package ids

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// UUID version and variant constants per RFC 9562.
const (
	uuidV7VersionMask = 0x0F
	uuidV7Version     = 0x70
	uuidV7VariantMask = 0x3F
	uuidV7Variant     = 0x80
)

// ErrInvalidUUIDLength is returned when a UUID string has an unexpected length.
var ErrInvalidUUIDLength = errors.New("invalid UUID length")

// UUID is a 16-byte universally unique identifier.
type UUID [16]byte

// NewV7 generates a UUIDv7 with the current time.
func NewV7() UUID {
	return NewV7WithTime(time.Now())
}

// NewV7WithTime generates a UUIDv7 with a specific timestamp.
func NewV7WithTime(t time.Time) UUID {
	var u UUID

	ms := uint64(max(t.UnixMilli(), 0)) //nolint:gosec // non-negative value guaranteed by max()

	// First 48 bits: milliseconds since epoch
	u[0] = byte(ms >> 40) //nolint:mnd // bit shift
	u[1] = byte(ms >> 32) //nolint:mnd // bit shift
	u[2] = byte(ms >> 24) //nolint:mnd // bit shift
	u[3] = byte(ms >> 16) //nolint:mnd // bit shift
	u[4] = byte(ms >> 8)  //nolint:mnd // bit shift
	u[5] = byte(ms)

	// Random bytes for the rest
	_, _ = rand.Read(u[6:])

	// Set version 7
	u[6] = (u[6] & uuidV7VersionMask) | uuidV7Version

	// Set variant 10
	u[8] = (u[8] & uuidV7VariantMask) | uuidV7Variant

	return u
}

// ParseUUID parses a hyphenated UUID string.
func ParseUUID(s string) (UUID, error) {
	s = strings.ReplaceAll(s, "-", "")

	if len(s) != 32 { //nolint:mnd // UUID hex length
		return UUID{}, fmt.Errorf("%w: %d", ErrInvalidUUIDLength, len(s))
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return UUID{}, fmt.Errorf("decode UUID hex: %w", err)
	}

	var u UUID
	copy(u[:], b)

	return u, nil
}

// String returns the hyphenated UUID string.
func (u UUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16],
	)
}

// IsZero returns true if the UUID is all zeros.
func (u UUID) IsZero() bool {
	return u == UUID{}
}

// Bytes returns the UUID as a byte slice.
func (u UUID) Bytes() []byte {
	b := make([]byte, 16) //nolint:mnd // UUID byte length
	copy(b, u[:])

	return b
}

// Compare returns -1, 0, or 1.
func (u UUID) Compare(other UUID) int {
	for i := range u {
		switch {
		case u[i] < other[i]:
			return -1
		case u[i] > other[i]:
			return 1
		}
	}

	return 0
}
