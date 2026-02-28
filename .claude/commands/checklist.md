---
description: Pre-commit code generation checklist with artifact status
---

# Pre-Commit Checklist

Before committing, verify each item:

## Code Quality
- [ ] Every function <= 70 lines and does one thing
- [ ] All inputs validated at boundaries
- [ ] All errors checked and wrapped with context
- [ ] No global mutable state
- [ ] `context.Context` is the first parameter where applicable
- [ ] Interfaces defined at the consumer, <= 3 methods
- [ ] No magic numbers (use named constants)

## Testing
- [ ] Tests exist for new behavior
- [ ] Table-driven with both valid and invalid cases
- [ ] Uses standard `testing` package (no testify)

## Verification
- [ ] `make test` passes → check `artifacts/test.log`
- [ ] `make test-all` passes → check `artifacts/test-all.log`
- [ ] `make lint` passes → check `artifacts/lint.log`
- [ ] No TODOs added without explicit approval
