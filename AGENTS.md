# General Guidelines

## Quick Reference

- Build: `go build ./cmd/dotgh`
- Test: `go test ./...`
- Lint: `golangci-lint run`

## Project Structure

- `docs/` - Documentation (concepts, design, development guide)
- `README.md` - Project overview, installation, usage

## Principles

- **YAGNI**: Don't implement until necessary.
- **KISS**: Avoid unnecessary complexity.
- **DRY**: Reuse through functions, methods, or packages.
- **Explicit errors**: Handle errors explicitly; avoid panics.
- **Small functions**: Prefer focused, single-purpose functions.

For details (SOLID, logging, etc.), see `docs/development.md`

## Testing

- **TDD**: Write tests before implementation.
- Use table-driven tests with `t.Run()`.

For testing patterns and examples, see `docs/development.md`

## Task Management

- `// TODO:` - Not yet implemented.
- `// FIXME:` - Broken code needing fix.
- `// HACK:` - Temporary workaround.
- Large tasks → GitHub Issue (`gh issue create`).

For workflow details, see `docs/development.md`
