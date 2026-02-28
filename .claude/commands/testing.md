---
description: Testing conventions, lint rules, make targets
---

# Testing Standards

## Conventions
- Standard `testing` package only — **no testify**.
- Use `if` + `t.Errorf` / `t.Fatalf` for assertions.
- Table-driven tests with `t.Run`.
- Compare structs with `google/go-cmp`.
- Separate success and error test cases.

## Test Structure
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "valid input", input: "hello", want: "HELLO"},
        {name: "empty input", input: "", wantErr: true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Feature(tc.input)
            if tc.wantErr {
                if err == nil { t.Fatal("expected error") }
                return
            }
            if err != nil { t.Fatalf("unexpected error: %v", err) }
            if got != tc.want { t.Errorf("got %q, want %q", got, tc.want) }
        })
    }
}
```

## Lint Rules
- `mnd` catches magic numbers — use named constants.
- `funcorder` requires unexported methods after exported methods.
- `funlen` max 80 lines per function.
- `gosec G602` flags array index expressions.

## Make Targets
- `make test` — unit tests → `artifacts/test.log`
- `make test-all` — integration tests (requires docker) → `artifacts/test-all.log`
- `make lint` — run linter → `artifacts/lint.log`
- `make coverage` — generate coverage reports
