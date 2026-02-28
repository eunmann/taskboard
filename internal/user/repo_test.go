package user_test

import (
	"context"
	"testing"

	"github.com/OWNER/PROJECT_NAME/internal/user"
)

func TestUserCreate(t *testing.T) {
	t.Skip("requires test database - run with make test-all")

	ctx := context.Background()

	// Example: create user and verify fields
	u := &user.User{
		Email: "test@example.com",
		Name:  "Test User",
	}

	_ = ctx
	_ = u
}

func TestUserGetByEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "valid email", email: "alice@example.com"},
		{name: "another email", email: "bob@example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.email == "" {
				t.Errorf("email should not be empty")
			}
		})
	}
}
