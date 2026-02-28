// Package main provides a database seeding tool for local development.
package main

import (
	"context"
	"flag"

	"github.com/OWNER/PROJECT_NAME/internal/platform/db"
	"github.com/OWNER/PROJECT_NAME/internal/platform/ids"
	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
	"github.com/OWNER/PROJECT_NAME/internal/user"
)

func main() {
	fresh := flag.Bool("fresh", false, "Reset all data before seeding")

	flag.Parse()

	logger := logging.SetupFromEnv()
	ctx := context.Background()

	logger.Info().Bool("fresh", *fresh).Msg("starting seeder")

	// Connect to database
	pool, err := db.ConnectFromEnv(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer pool.Close()

	// Fresh: truncate tables
	if *fresh {
		logger.Warn().Msg("truncating all tables")

		_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to truncate tables")
		}
	}

	// Seed users
	userRepo := user.NewRepo(pool)

	seedUsers := []struct {
		email string
		name  string
	}{
		{email: "alice@example.com", name: "Alice Johnson"},
		{email: "bob@example.com", name: "Bob Smith"},
		{email: "charlie@example.com", name: "Charlie Brown"},
	}

	for _, su := range seedUsers {
		existing, _ := userRepo.GetByEmail(ctx, su.email)
		if existing != nil {
			logger.Info().Str("email", su.email).Msg("user already exists, skipping")

			continue
		}

		u := &user.User{
			ID:    ids.NewUserID(),
			Email: su.email,
			Name:  su.name,
		}

		if err := userRepo.Create(ctx, u); err != nil {
			logger.Error().Err(err).Str("email", su.email).Msg("failed to create user")

			continue
		}

		logger.Info().Str("email", su.email).Str("id", u.ID.String()).Msg("created user")
	}

	logger.Info().Msg("seeding complete")
}
