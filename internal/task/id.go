package task

import (
	"crypto/rand"
	"strings"
)

// idLen is the length of a generated task ID (6 characters per spec §4.1).
const idLen = 6

// alphabetSize is the number of characters in the ID alphabet.
const alphabetSize = 54

// alphabet excludes visually ambiguous characters: 0, O, o, 1, l, L, I, i.
const alphabet = "23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"

// NewID generates a random 6-character base-56 task ID.
func NewID() string {
	b := make([]byte, idLen)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}

	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}

	return string(b)
}

// ValidID reports whether s is a valid task ID (correct length, valid characters).
func ValidID(s string) bool {
	if len(s) != idLen {
		return false
	}

	for _, c := range s {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}

	return true
}
