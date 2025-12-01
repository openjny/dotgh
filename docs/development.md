# Development Guide

## Testing Strategy

### Tools & Frameworks

- Use Go's standard `testing` package with `testify/assert` for assertions.

### Table-Driven Tests

Define test cases as a slice of structs with `name`, inputs, and expected outputs.

```go
tests := []struct {
    name     string
    input    string
    expected int
}{
    {"empty string", "", 0},
    {"single word", "hello", 5},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := myFunc(tt.input)
        assert.Equal(t, tt.expected, result)
    })
}
```

### Testability Patterns

- **Constructor pattern**: Use `NewXxxCmd()` functions instead of `init()` for testability.
- **Dependency injection**: Use interfaces for mocking external dependencies.
- **Interface-based fakes**: Prefer simple fakes over mock frameworks (`gomock`) for simple interfaces.

## Coding Principles (Details)

### SOLID Principles

- **S**ingle Responsibility: Each module/class should have one reason to change.
- **O**pen/Closed: Open for extension, closed for modification.
- **L**iskov Substitution: Subtypes must be substitutable for their base types.
- **I**nterface Segregation: Many specific interfaces are better than one general-purpose interface.
- **D**ependency Inversion: Depend on abstractions, not concretions.

### Error Handling

- Handle errors explicitly. Avoid panics except for unrecoverable situations.
- Wrap errors with context using `fmt.Errorf("context: %w", err)`.

### Logging

- Use structured logging for better log management and analysis.
- Include relevant context (request ID, user ID, etc.) in log entries.

## Task Management Workflow

### GitHub Issues

Use `gh issue` to track features, bugs, and tasks that take more than a few hours.

### Code Comments

For small, localized tasks within a PR scope:

- `// TODO:` - Feature not yet implemented.
- `// FIXME:` - Broken or incorrect code that needs fixing.
- `// HACK:` - Temporary workaround; should be replaced with a proper solution.

If a code TODO grows beyond a quick fix, convert it to a GitHub Issue.

### AI-Generated Code

- Do not blindly accept AI-generated TODOs. Verify if the task is necessary and actionable.
- Review AI suggestions critically before committing.
