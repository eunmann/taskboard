---
description: Branch creation, incremental change process, PR steps
---

# Workflow

## Starting Work

1. **Check current state**: `git status` and `git log --oneline -5`
2. **Create feature branch**: `git checkout -b feature/<name>` from `main`
3. **Understand the task**: Read relevant files before writing code

## Making Changes

1. **Small, atomic commits** — one logical change per commit
2. **Run tests after each change**: `make test`
3. **Run linter**: `make lint`
4. **Commit with clear message**: describe the "why", not the "what"

## Finishing

1. **Run full test suite**: `make test-all`
2. **Run linter**: `make lint`
3. **Check for gaps**: review all changed files
4. **Push and create PR**: `git push -u origin HEAD`
