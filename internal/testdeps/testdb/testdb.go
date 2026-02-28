// Package testdb provides per-test database isolation for integration tests.
package testdb

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/platform/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB wraps a pgxpool.Pool with per-test isolation.
type TestDB struct {
	Pool      *pgxpool.Pool
	DBName    string
	adminPool *pgxpool.Pool
}

// New creates an isolated test database cloned from the template.
func New(t *testing.T) *TestDB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:mnd // test constant
	defer cancel()

	// Build admin DSN (connects to postgres database)
	host := env.GetOrDefault(env.PostgresHost, "localhost")
	port := env.GetOrDefault(env.PostgresPort, "5433")
	user := env.GetOrDefault(env.PostgresUser, "platform")
	pass := env.GetOrDefault(env.PostgresPassword, "platform_dev")
	templateDB := env.GetOrDefault(env.PostgresDB, "platform_test")

	adminDSN := fmt.Sprintf("postgres://%s:%s@%s/postgres?sslmode=disable", user, pass, net.JoinHostPort(host, port))

	adminPool, err := pgxpool.New(ctx, adminDSN)
	if err != nil {
		t.Fatalf("connect admin pool: %v", err)
	}

	// Create unique database name
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	safeName = strings.ReplaceAll(safeName, " ", "_")
	dbName := fmt.Sprintf("test_%s_%d", safeName, time.Now().UnixNano())

	// Truncate to PostgreSQL max identifier length
	maxLen := 63
	if len(dbName) > maxLen {
		dbName = dbName[:maxLen]
	}

	// Create database as clone of template
	query := fmt.Sprintf(
		`CREATE DATABASE %q TEMPLATE %q OWNER %q`,
		dbName, templateDB, user,
	)

	_, err = adminPool.Exec(ctx, query)
	if err != nil {
		adminPool.Close()
		t.Fatalf("create test database: %v", err)
	}

	// Connect to the new test database
	testDSN := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, net.JoinHostPort(host, port), dbName)

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		adminPool.Close()
		t.Fatalf("connect test pool: %v", err)
	}

	tdb := &TestDB{
		Pool:      pool,
		DBName:    dbName,
		adminPool: adminPool,
	}

	t.Cleanup(func() {
		pool.Close()

		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // test constant
		defer cleanCancel()

		// Terminate connections and drop database
		terminateQuery := fmt.Sprintf(
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid != pg_backend_pid()`,
			dbName,
		)
		_, _ = adminPool.Exec(cleanCtx, terminateQuery)

		dropQuery := fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, dbName)
		_, _ = adminPool.Exec(cleanCtx, dropQuery)

		adminPool.Close()
	})

	return tdb
}

// Exec executes a query on the test database.
func (tdb *TestDB) Exec(t *testing.T, query string, args ...any) {
	t.Helper()

	ctx := context.Background()

	_, err := tdb.Pool.Exec(ctx, query, args...)
	if err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
