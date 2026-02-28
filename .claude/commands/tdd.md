---
description: Red-Green-Refactor test-driven development cycle
---

# TDD — Red-Green-Refactor

## Cycle

### 1. RED — Write a Failing Test
- Write the test first, before any implementation.
- Run the test — it must fail.
- The test defines the expected behavior.

### 2. GREEN — Make It Pass
- Write the minimum code to make the test pass.
- No extra features. No optimization. Just pass.
- Run the test — it must pass.

### 3. REFACTOR — Clean Up
- Improve code quality without changing behavior.
- Follow `/go-standards`.
- Run tests again — still passing.

## Rules

- **One test per cycle.** Don't write multiple failing tests.
- **Small cycles.** Each cycle should take minutes, not hours.
- **Commit after green.** One test + implementation per commit.
- **Don't refactor while red.** Get green first, then refactor.

## Process

```
1. Write test          → make test (FAIL)
2. Write minimal code  → make test (PASS)
3. Refactor            → make test (PASS)
4. Commit              → git commit
5. Repeat
```
