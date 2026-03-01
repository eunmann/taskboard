---
description: Per-task implementation verification procedure
---

# Verify Task

Run this after completing a task to verify it meets all quality standards.

## 1. Identify Scope

- What files were changed? `git diff --name-only HEAD~1`
- What behavior was added or modified?
- Were tests added for all new behavior?
- Confirm task status: `grep "^status:" .taskboard/<id>-*.yaml`

## 2. Run Tests

```
make test
```

- Check `artifacts/test.log` for failures.
- If any test fails, fix the code (not the test, unless the test is wrong).
- Verify coverage is reasonable for the changed code.

## 3. Run Linter

```
make lint
```

- Check `artifacts/lint.log` for warnings.
- Zero warnings required. Fix by improving code, not suppressing.
- Verify formatting is applied (linter auto-fixes with `--fix`).

## 4. Review Changes

For each changed file, verify:

- [ ] Every function <= 70 lines
- [ ] All errors checked and wrapped with context
- [ ] No global mutable state introduced
- [ ] No magic numbers (use named constants)
- [ ] No TODOs added without approval
- [ ] Exported types and functions have doc comments

## 5. Check Consistency

- [ ] New code follows existing patterns in the package
- [ ] Naming is consistent with surrounding code
- [ ] No unnecessary dependencies added

## 6. Confirm

If all checks pass:
- Commit with a clear message describing the "why"
- Mark the task as complete
- Pick up the next task
