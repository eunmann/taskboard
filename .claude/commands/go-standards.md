---
description: Full Go code quality standards (12 subsections)
---

# Go Standards

## 1. Read First, Write Second
Always read existing code before modifying. Understand patterns before changing them.

## 2. Control Flow
- No goto. Use early returns to reduce nesting.
- Guard clauses at the top of functions.

## 3. Bounds and Allocation
- Fixed bounds on loops and data structures.
- Pre-allocate slices when size is known: `make([]T, 0, n)`.

## 4. Functions
- Maximum 70 lines per function (lint limit: 80).
- Each function does one thing.
- 3-4 parameters maximum; use a struct for more.

## 5. Error Handling
- Check every error. Never use `_` for errors.
- Wrap with context: `fmt.Errorf("operation: %w", err)`.
- Use sentinel errors for expected conditions.

## 6. Types and Interfaces
- Accept interfaces, return structs.
- Interfaces: 1-3 methods, defined at the consumer.

## 7. Scope and State
- Smallest possible scope for variables.
- No global mutable state.

## 8. Concurrency
- Clear ownership of data.
- Use `sync.WaitGroup` for goroutine coordination.
- `context.Context` as the first parameter.

## 9. Pointers and Unsafe
- Avoid pointers unless mutation is needed.
- Never use `unsafe`.

## 10. Naming
- MixedCaps (no underscores in Go names).
- Short receiver names (1-2 chars).
- No package name stuttering.

## 11. Documentation
- Doc comments on all exported types and functions.

## 12. Architecture
- Single-purpose packages.
- Thin `main()` — delegate to packages.
- DAG dependency structure (no cycles).
