// Package user provides user management and persistence.
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/platform/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a platform user.
type User struct {
	ID        ids.UserID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repo provides user persistence operations.
type Repo struct {
	pool *pgxpool.Pool
}

// NewRepo creates a new user repository.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Create inserts a new user.
func (r *Repo) Create(ctx context.Context, u *User) error {
	if u.ID.IsZero() {
		u.ID = ids.NewUserID()
	}

	query := `INSERT INTO users (id, email, name) VALUES ($1, $2, $3) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, u.ID, u.Email, u.Name).
		Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *Repo) GetByID(ctx context.Context, id ids.UserID) (*User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`

	var u User

	err := r.pool.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

// GetByEmail retrieves a user by email.
func (r *Repo) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1`

	var u User

	err := r.pool.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

// List returns all users, ordered by creation time.
func (r *Repo) List(ctx context.Context) ([]User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	defer rows.Close()

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (User, error) {
		var u User

		if scanErr := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); scanErr != nil {
			return u, fmt.Errorf("scan user: %w", scanErr)
		}

		return u, nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect users: %w", err)
	}

	return users, nil
}
