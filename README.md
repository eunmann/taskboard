# PROJECT_NAME

A Go web application template with HTMX, Alpine.js, Tailwind CSS, PostgreSQL, and AWS CDK.

## Using This Template

### Automated Setup (Recommended)

1. Click **"Use this template"** on GitHub (or clone and remove `.git/`)
2. Run the init script:
   ```bash
   ./init.sh <your-github-org> <your-project-name>
   ```
3. Review and commit the changes
4. Delete `init.sh`
5. Follow the Quick Start below

### Manual Setup

Replace these placeholders across the codebase:

| Find | Replace With |
|------|-------------|
| `OWNER` | Your GitHub username or org |
| `PROJECT_NAME` | Your project name |
| `app-storage` | Your S3 bucket prefix |

Then run `go mod tidy` in both the root and `cdk/` directories.

### After Initialization

- Update CDK config (`cdk/config/environments.go`) with your VPC CIDR, RDS instance class, domain, and site URL
- Update this README to describe your project
- Delete `init.sh`

## Prerequisites

- [Go 1.25+](https://golang.org/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Node.js](https://nodejs.org/) (for AWS CDK CLI)
- [AWS CDK CLI](https://docs.aws.amazon.com/cdk/latest/guide/cli.html) (for infrastructure deployment)
- [jq](https://jqlang.github.io/jq/) (used by Claude Code pre-commit hooks)

## Quick Start

```bash
# 1. Start infrastructure (Postgres, MinIO, ElasticMQ, MailHog)
make dev-up

# 2. Run database migrations
make migrate-up

# 3. Seed sample data
make seed

# 4. Open http://localhost:8080
```

The backend runs inside Docker with hot-reload via [air](https://github.com/air-verse/air).

## Running Locally (without Docker backend)

```bash
# Start only infrastructure services
docker compose -p app -f infra/docker-compose.yml up -d

# Run migrations
make migrate-up

# Seed data
make seed

# Run the dev server directly
go run ./services/dev
```

## Running Tests

```bash
# Unit tests only
make test

# Integration tests (requires docker services running)
make dev-up
make db-reset-test
make test-all

# View results
grep "FAIL" artifacts/test.log
grep "FAIL" artifacts/test-all.log
```

## Linting

```bash
make lint          # Auto-fix and report
make lint-check    # CI mode (no auto-fix)
```

## Database Migrations

```bash
make migrate-up                       # Apply pending migrations
make migrate-down                     # Rollback last migration
make migrate-create name=add_column   # Create new migration
make migrate-version                  # Show current version
```

## Deployment (AWS CDK)

```bash
# First-time setup
make cdk-deps       # Install CDK CLI
make cdk-bootstrap  # Bootstrap CDK in AWS account

# Deploy
make cdk-deploy-network   # VPC, security groups
make cdk-deploy-stateful  # RDS, S3, SQS
make cdk-deploy-compute   # ECS cluster, backend service

# Or deploy everything
make cdk-deploy-all
```

## Project Structure

```
.
├── .claude/                    # Claude Code configuration & skills
│   ├── settings.json          # Pre-commit lint hook
│   ├── hooks/                 # Git hooks
│   └── commands/              # Slash commands (workflow, testing, etc.)
├── .github/workflows/         # GitHub Actions CI/CD
├── cdk/                       # AWS CDK infrastructure (Go)
│   ├── config/                # Environment configuration
│   ├── stacks/                # Stack definitions
│   └── main.go                # CDK app entry point
├── infra/                     # Docker & database
│   ├── docker-compose.yml     # Dev + test services
│   ├── Dockerfile.dev         # Hot-reload dev container
│   ├── Dockerfile.backend     # Multi-stage production build
│   ├── elasticmq.conf         # SQS queue configuration
│   └── migrations/            # PostgreSQL migration files
├── internal/                  # Core application code
│   ├── app/                   # DI container, config, service lifecycle
│   ├── httpapi/               # HTTP router, middleware, response helpers
│   ├── platform/              # Infrastructure abstractions
│   │   ├── db/                # PostgreSQL connection & transactions
│   │   ├── env/               # Environment variable utilities
│   │   ├── ids/               # Type-safe UUID identifiers
│   │   └── logging/           # Structured logging (zerolog)
│   ├── user/                  # User domain (repo, model)
│   ├── web/                   # Web UI handlers & templates
│   │   ├── static/            # CSS, JS, vendor libraries
│   │   ├── templates/         # HTML templates (layouts, pages, partials)
│   │   └── testhelpers/       # goquery-based HTML test assertions
│   └── testdeps/              # Test infrastructure (isolated DB)
├── services/                  # Application entry points
│   ├── backend/               # Production HTTP server
│   └── dev/                   # All-in-one dev runner
├── tools/                     # CLI utilities
│   └── seed/                  # Database seeder
├── artifacts/                 # Build outputs (test.log, lint.log, coverage/)
├── .air.toml                  # Hot-reload configuration
├── .golangci.yml              # Linter configuration
├── .env.dev                   # Development environment
├── .env.test                  # Test environment
├── .env.ci                    # CI environment
├── CLAUDE.md                  # Project standards & conventions
├── Makefile                   # Developer workflow automation
└── go.mod                     # Go module definition
```

## Available Make Targets

| Target | Description |
|--------|-------------|
| `make help` | Show all targets |
| `make build` | Build all binaries |
| `make clean` | Remove build artifacts |
| `make dev-up` | Start dev infrastructure |
| `make dev-down` | Stop infrastructure |
| `make dev-reset` | Stop and remove volumes |
| `make dev-logs` | Follow backend logs |
| `make seed` | Seed development data |
| `make seed-fresh` | Reset and re-seed (destructive) |
| `make migrate-up` | Run pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create` | Create new migration |
| `make test` | Run unit tests |
| `make test-all` | Run integration tests |
| `make coverage` | Generate coverage reports |
| `make lint` | Run linter with auto-fix |
| `make lint-check` | Run linter (CI mode) |
| `make cdk-synth` | Synthesize CloudFormation |
| `make cdk-diff` | Show infrastructure changes |
| `make cdk-deploy-all` | Deploy all stacks |

## Tech Stack

- **Backend**: Go 1.25, Chi router, pgx PostgreSQL driver, zerolog
- **Frontend**: HTMX, Alpine.js, Tailwind CSS (Play CDN)
- **Database**: PostgreSQL 18, golang-migrate
- **Infrastructure**: AWS CDK (Go), ECS Fargate, RDS, S3, SQS
- **Testing**: Standard `testing` package, goquery HTML assertions
- **CI/CD**: GitHub Actions
- **Dev Tools**: Docker Compose, air hot-reload, golangci-lint v2
