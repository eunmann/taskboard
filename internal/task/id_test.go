package task

import (
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != idLen {
		t.Errorf("NewID() length = %d, want %d", len(id), idLen)
	}

	if !ValidID(id) {
		t.Errorf("NewID() = %q, not valid", id)
	}
}

func TestNewIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for range 100 {
		id := NewID()
		if seen[id] {
			t.Fatalf("duplicate ID generated: %q", id)
		}

		seen[id] = true
	}
}

func TestValidID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "Abc456", true},
		{"too short", "abc", false},
		{"too long", "abc1234", false},
		{"invalid char 0", "0bc123", false},
		{"invalid char O", "Obc123", false},
		{"invalid char o", "obc123", false},
		{"invalid char l", "lbc123", false},
		{"invalid char L", "Lbc123", false},
		{"invalid char 1", "1bc123", false},
		{"invalid char I", "Ibc123", false},
		{"invalid char i", "ibc123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidID(tt.input); got != tt.want {
				t.Errorf("ValidID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAlphabetLength(t *testing.T) {
	if len(alphabet) != alphabetSize {
		t.Errorf("alphabet length = %d, want %d", len(alphabet), alphabetSize)
	}
}

func TestAlphabetExcludesAmbiguous(t *testing.T) {
	excluded := "0OoI1ilL"
	for _, c := range excluded {
		if containsRune(alphabet, c) {
			t.Errorf("alphabet contains excluded character %q", string(c))
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}

	return false
}
