---
description: Branch creation, incremental change process, PR steps
---

# Workflow

## 1. Understand

- Read the task requirements carefully.
- Read relevant code before writing any changes.
- Identify which files and packages are affected.

## 2. Branch

- Check current state: `git status` and `git log --oneline -5`
- Create feature branch: `git checkout -b feature/<name>` from `main`

## 3. Implement

Follow the Execution Loop (see CLAUDE.md):

1. Decompose into small, testable steps.
2. Implement one step at a time — write tests first when adding behavior (`/tdd`).
3. Verify after each step: `make test`, `make lint`.
4. Commit atomically — one logical change per commit.
5. Run `/verify-task` after each meaningful step.

## 4. Verify

- `make test` — check `artifacts/test.log`
- `make test-all` — check `artifacts/test-all.log`
- `make lint` — check `artifacts/lint.log`
- All must pass with zero failures and zero warnings.

## 5. Review

- Run `/review-branch` to check all changes for gaps.
- Fix any issues found before proceeding.

## 6. Ship

- Push: `git push -u origin HEAD`
- Create PR with clear title and description.
- Never push broken builds.
