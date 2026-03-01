---
description: Pre-commit code generation checklist with artifact status
---

# Pre-Commit Checklist

## 1. Scan

Review all staged changes:
- `git diff --cached --stat` — list changed files
- Read each changed file completely

## 2. Verify Code Quality

For each changed file, check:
- [ ] Every function <= 70 lines and does one thing
- [ ] All inputs validated at boundaries
- [ ] All errors checked and wrapped with context
- [ ] No global mutable state
- [ ] `context.Context` is the first parameter where applicable
- [ ] Interfaces defined at the consumer, <= 3 methods
- [ ] No magic numbers (use named constants)

## 3. Verify Tests

- [ ] Tests exist for every new behavior
- [ ] Table-driven with both valid and invalid cases
- [ ] Uses standard `testing` package (no testify)
- [ ] HTML assertions use `goquery` selectors where applicable

## 4. Run Verification

```
make test      → check artifacts/test.log
make test-all  → check artifacts/test-all.log
make lint      → check artifacts/lint.log
```

- All three must pass with zero failures and zero warnings.
- If any fail, fix the issue before proceeding.

## 5. Commit

- [ ] No TODOs added without explicit approval
- [ ] Commit message describes the "why", not the "what"
- [ ] One logical change per commit
