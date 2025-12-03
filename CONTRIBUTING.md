# Contributing to dotgh

Thank you for your interest in contributing to dotgh! This document provides guidelines and information for contributors.

## Language Policy

- **Code and comments**: English only
- **Documentation** (`docs/`, `README.md`): English
- **Commit messages**: English (Conventional Commits style)
- **Issues and Pull Requests**: English

## Getting Started

### Prerequisites

- Go 1.24 or later
- Make
- Git

### Setup

```bash
git clone https://github.com/openjny/dotgh.git
cd dotgh
make build
```

### Running Tests

```bash
make test              # Run all unit and integration tests
make test-short        # Run only unit tests (skip integration)
make test-e2e          # Build binary and run E2E tests
make test-cover        # Run tests with coverage report
```

## Development Workflow

### Branch Strategy

- **Never commit directly to `main`**
- Create a feature branch for each task: `feature/issue-{number}-description`
- Keep branches up to date by rebasing or merging from `main`

### Commit Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Example:**
```
feat: add config validation for includes patterns

Validate that includes patterns are valid glob syntax before saving.

Closes #42
```

### Pull Request Process

1. Create a feature branch from `main`
2. Make your changes following the coding guidelines
3. Ensure all tests pass (`make test`)
4. Push your branch and create a Pull Request
5. Reference related issues (e.g., `Closes #123`)

## Coding Guidelines

### Principles

- **YAGNI**: Don't implement until necessary
- **KISS**: Avoid unnecessary complexity
- **DRY**: Reuse through functions, methods, or packages
- **Explicit errors**: Handle errors explicitly; avoid panics
- **Small functions**: Prefer focused, single-purpose functions

### Code Style

- Run `go fmt` before committing
- Run `go vet` to catch common mistakes
- Use meaningful variable and function names

### Testing

- **TDD**: Write tests before implementation when possible
- Use table-driven tests with `t.Run()`
- Target ratio: 70% unit, 20% integration, 10% e2e tests

For detailed testing patterns, see [docs/development.md](docs/development.md).

## Task Management

### Code Comments

- `// TODO:` - Feature not yet implemented
- `// FIXME:` - Broken code needing fix
- `// HACK:` - Temporary workaround

### GitHub Issues

For larger tasks that take more than a few hours, create a GitHub Issue:

```bash
gh issue create --title "feat: your feature" --body "Description..."
```

## Questions?

Feel free to open an issue for questions or discussions.
