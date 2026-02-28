---
description: Review all branch changes, find gaps, fix them
---

# Review Branch

## Process

### 1. Gather Context
- `git diff main...HEAD --stat` — see all changed files
- `git log main..HEAD --oneline` — see all commits
- Read every changed file completely

### 2. Check for Gaps

For each changed file, check:
- [ ] **Tests**: Does every new behavior have a test?
- [ ] **Edge cases**: Are error paths and boundary conditions covered?
- [ ] **Error handling**: Are all errors checked and wrapped?
- [ ] **Function length**: Is every function <= 70 lines?
- [ ] **Lint**: Does `make lint` pass with zero warnings?
- [ ] **Documentation**: Are exported types documented?

### 3. Fix Issues
- Add missing tests
- Refactor oversized functions
- Wrap unwrapped errors
- Fix lint issues by improving code (not suppressing)

### 4. Report

Create a summary table:

| File | Issue | Status |
|------|-------|--------|
| internal/foo/bar.go | Missing error test | Fixed |
| internal/baz/qux.go | Function too long (85 lines) | Refactored |

Then run:
- `make test` — verify unit tests pass
- `make test-all` — verify integration tests pass
- `make lint` — verify zero lint warnings
