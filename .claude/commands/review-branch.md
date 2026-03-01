---
description: Review all branch changes, find gaps, fix them
---

# Review Branch

## 1. Gather Context

- `git diff main...HEAD --stat` — see all changed files
- `git log main..HEAD --oneline` — see all commits
- Read every changed file completely

## 2. Check Each File

For each changed file, verify:

- [ ] **Tests**: Does every new behavior have a test?
- [ ] **Edge cases**: Are error paths and boundary conditions covered?
- [ ] **Error handling**: Are all errors checked and wrapped?
- [ ] **Function length**: Is every function <= 70 lines?
- [ ] **Lint**: Does `make lint` pass with zero warnings?
- [ ] **Documentation**: Are exported types documented?

## 3. Check Consistency

- [ ] New code follows existing patterns in the package
- [ ] Naming is consistent with surrounding code
- [ ] No unnecessary dependencies added
- [ ] No dead code or unused imports

## 4. Fix Issues

- Add missing tests
- Refactor oversized functions
- Wrap unwrapped errors
- Fix lint issues by improving code (not suppressing)

## 5. Verify

Run all checks and confirm clean results:

- `make test` — check `artifacts/test.log`
- `make test-all` — check `artifacts/test-all.log`
- `make lint` — check `artifacts/lint.log`

## 6. Report

Create a summary table:

| File | Issue | Status |
|------|-------|--------|
| internal/foo/bar.go | Missing error test | Fixed |
| internal/baz/qux.go | Function too long (85 lines) | Refactored |

If all checks pass and no issues remain, the branch is ready to ship.
