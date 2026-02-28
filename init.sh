#!/usr/bin/env bash
set -euo pipefail

# Template initialization script
# Replaces OWNER/PROJECT_NAME placeholders throughout the codebase.
#
# Usage: ./init.sh <github-owner> <project-name>
# Example: ./init.sh acme my-saas-app

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <github-owner> <project-name>"
    echo "Example: $0 acme my-saas-app"
    exit 1
fi

OWNER="$1"
PROJECT="$2"

# Capitalize first letter for CDK prefixes (App- -> Myproject-)
PROJECT_CAPITALIZED="$(echo "${PROJECT:0:1}" | tr '[:lower:]' '[:upper:]')${PROJECT:1}"

# Guard: refuse to run if placeholders are already replaced
if ! grep -q 'github.com/OWNER/PROJECT_NAME' go.mod; then
    echo "Error: Placeholders already replaced. This script should only run once."
    exit 1
fi

# Cross-platform sed -i (GNU vs BSD/macOS)
sedi() {
    if sed --version >/dev/null 2>&1; then
        # GNU sed
        sed -i "$@"
    else
        # BSD sed (macOS)
        sed -i '' "$@"
    fi
}

echo "Initializing project: github.com/${OWNER}/${PROJECT}"
echo ""

# 1. Replace Go module path in go.mod files
echo "Replacing module paths..."
sedi "s|github.com/OWNER/PROJECT_NAME|github.com/${OWNER}/${PROJECT}|g" go.mod
sedi "s|github.com/OWNER/PROJECT_NAME|github.com/${OWNER}/${PROJECT}|g" cdk/go.mod

# 2. Replace import paths in all Go source files
echo "Replacing Go import paths..."
find . -name '*.go' -not -path './vendor/*' -exec \
    grep -l 'github.com/OWNER/PROJECT_NAME' {} + 2>/dev/null | while read -r file; do
    sedi "s|github.com/OWNER/PROJECT_NAME|github.com/${OWNER}/${PROJECT}|g" "$file"
done

# 3. Replace in linter config
echo "Replacing linter config..."
sedi "s|github.com/OWNER/PROJECT_NAME|github.com/${OWNER}/${PROJECT}|g" .golangci.yml

# 4. Replace in Makefile
echo "Replacing Makefile references..."
sedi "s|github.com/OWNER/PROJECT_NAME|github.com/${OWNER}/${PROJECT}|g" Makefile
sedi "s|^# PROJECT_NAME|# ${PROJECT}|" Makefile
sedi "s|^PROJECT := app|PROJECT := ${PROJECT}|" Makefile

# 5. Replace README title
echo "Replacing README title..."
sedi "s|^# PROJECT_NAME|# ${PROJECT}|" README.md

# 6. Replace S3 bucket names (app-storage -> <project>-storage)
echo "Replacing S3 bucket names..."
for f in .env.dev .env.test .env.ci; do
    sedi "s|app-storage|${PROJECT}-storage|g" "$f"
done
sedi "s|app-storage|${PROJECT}-storage|g" infra/docker-compose.yml
sedi "s|app-storage|${PROJECT}-storage|g" .github/workflows/pr-validation.yml

# 7. Replace CDK config prefixes (App- -> <Project>-, app- -> <project>-)
echo "Replacing CDK config prefixes..."
sedi "s|\"App-%s-%s\"|\"${PROJECT_CAPITALIZED}-%s-%s\"|g" cdk/config/config.go
sedi "s|\"app-%s-%s\"|\"${PROJECT}-%s-%s\"|g" cdk/config/config.go
sedi "s|\"app-%s-%s-%s\"|\"${PROJECT}-%s-%s-%s\"|g" cdk/config/config.go

# 8. Replace CDK main.go references
echo "Replacing CDK main.go references..."
sedi "s|jsii.String(\"App\")|jsii.String(\"${PROJECT_CAPITALIZED}\")|g" cdk/main.go
sedi "s|App-\$(CDK_ENV)|${PROJECT_CAPITALIZED}-\$(CDK_ENV)|g" Makefile

# 9. Run go mod tidy
echo ""
echo "Running go mod tidy..."
go mod tidy
(cd cdk && go mod tidy)

echo ""
echo "Done! Project initialized as github.com/${OWNER}/${PROJECT}"
echo ""
echo "Next steps:"
echo "  1. Update cdk/config/environments.go with your VPC CIDR, RDS instance class, domain, and site URL"
echo "  2. Update README.md to describe your project"
echo "  3. Delete this script: rm init.sh"
echo "  4. Commit: git add -A && git commit -m 'Initialize project from template'"
