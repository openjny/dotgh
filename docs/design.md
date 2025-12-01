# dotgh Design

## Architecture

Implemented as a single-binary CLI application in Go.

### Tech Stack

- **Language**: Go
- **CLI Framework**: `spf13/cobra`
- **Self-Update**: `creativeprojects/go-selfupdate`
- **Release**: `goreleaser` v2 (integrated with GitHub Actions)

## Directory Structure

Complies with the XDG Base Directory Specification.

### Configuration & Template Storage Location

- **Linux/macOS**: `~/.config/dotgh/templates/`
- **Windows**: `~/AppData/Local/dotgh/templates/` (or unified to `~/.config/dotgh/templates/`)

### Repository Structure

```
dotgh/
├── cmd/dotgh/            # Entry point
├── internal/
│   ├── commands/         # CLI subcommands
│   ├── updater/          # Self-update logic
│   └── version/          # Version info (ldflags)
├── docs/                 # Documentation
└── .goreleaser.yaml      # Release configuration
```

## CLI Interface

```bash
dotgh [command] [flags]
```

### Command List

| Command   | Arguments    | Options           | Description                                         | Status      |
| --------- | ------------ | ----------------- | --------------------------------------------------- | ----------- |
| `list`    | None         | None              | Display a list of available templates               | Implemented |
| `apply`   | `<template>` | `-f, --force`     | Apply a template to the current directory           | Implemented |
| `push`    | `<template>` | `-f, --force`     | Save the current directory's settings as a template | Implemented |
| `delete`  | `<template>` | `-f, --force`     | Delete a template                                   | Implemented |
| `update`  | None         | `-c, --check`     | Update dotgh itself to the latest version           | Implemented |
| `version` | None         | None              | Display version information                         | Implemented |

## Template Targets

Default files/directories managed as template components:

- `.github/`
- `.vscode/`
- `AGENTS.md`

## Command Behavior

- **apply/push**: Use `-f` flag to overwrite existing files; without it, existing files are skipped.
- **update**: Downloads from GitHub Releases with checksum validation; skips if running dev build.
