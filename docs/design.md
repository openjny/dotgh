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
│   ├── config/           # Configuration management
│   ├── diff/             # File difference calculation
│   ├── editor/           # Editor detection and launching
│   ├── glob/             # Glob pattern matching
│   ├── prompt/           # User confirmation prompts
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

| Command        | Arguments    | Options                 | Description                                         | Status      |
| -------------- | ------------ | ----------------------- | --------------------------------------------------- | ----------- |
| `list`         | None         | None                    | Display a list of available templates               | Implemented |
| `pull`         | `<template>` | `-m, --merge`, `-y, --yes` | Pull a template to the current directory         | Implemented |
| `push`         | `<template>` | `-m, --merge`, `-y, --yes` | Save the current directory's settings as a template | Implemented |
| `diff`         | `<template>` | `-r, --reverse`, `--merge` | Show differences between template and current directory | Implemented |
| `delete`       | `<template>` | `-f, --force`           | Delete a template                                   | Implemented |
| `edit`         | `[template]` | `-c, --create`          | Open template in the user's preferred editor        | Implemented |
| `update`       | None         | `-c, --check`           | Update dotgh itself to the latest version           | Implemented |
| `version`      | None         | None                    | Display version information                         | Implemented |
| `config`       | None         | None                    | Manage dotgh configuration (parent command)         | Implemented |
| `config show`  | None         | None                    | Display current configuration in YAML format        | Implemented |
| `config edit`  | None         | None                    | Open configuration file in the user's preferred editor | Implemented |

## Template Targets

Default files/directories managed as template components (glob patterns):

- `AGENTS.md`
- `.github/agents/*.agent.md`
- `.github/copilot-chat-modes/*.chatmode.md`
- `.github/copilot-instructions.md`
- `.github/instructions/*.instructions.md`
- `.github/prompts/*.prompt.md`
- `.vscode/mcp.json`

### Configuration File

Targets can be customized via `~/.config/dotgh/config.yaml`:

```yaml
editor: "code --wait"  # Optional: override the default editor
templates_dir: "~/my-templates"  # Optional: custom templates directory location
includes:
  - "AGENTS.md"
  - ".github/agents/*.agent.md"
  - ".github/copilot-chat-modes/*.chatmode.md"
  - ".github/copilot-instructions.md"
  - ".github/instructions/*.instructions.md"
  - ".github/prompts/*.prompt.md"
  - ".vscode/mcp.json"
  - "custom/*.md"  # Add custom patterns
excludes:  # Optional: exclude specific files from template management
  - ".github/prompts/local.prompt.md"
  - ".github/prompts/secret-*.prompt.md"
```

If the config file doesn't exist, default targets are used.

### Exclude Pattern Behavior

- Files matching `excludes` patterns are filtered out after `includes` expansion
- `excludes` takes priority: if a file matches both `includes` and `excludes`, it is excluded
- Default: empty list (no exclusions)
- Supports the same glob pattern syntax as `includes`

### Editor Detection

The editor for `edit` and `config edit` commands is determined in the following order:

1. `editor` field in `config.yaml` (highest priority)
2. `VISUAL` environment variable
3. `EDITOR` environment variable
4. `GIT_EDITOR` environment variable
5. Platform-specific fallback: `vi` (Linux/macOS), `notepad` (Windows)

For GUI editors (VS Code, Sublime Text), the `--wait` flag is automatically added.

## Command Behavior

### pull/push (Git-style Sync)

By default, `pull` and `push` perform a **full sync** with Git-style behavior:

- **Adds** new files from source
- **Updates** modified files (overwrites with source content)
- **Deletes** files that exist in destination but not in source

This ensures the destination matches the source exactly.

**Flags:**
- `-m, --merge`: Merge mode - only add and update files, no deletions
- `-y, --yes`: Skip confirmation prompt

**Examples:**
```bash
dotgh pull my-template          # Full sync with confirmation
dotgh pull my-template --yes    # Full sync without confirmation
dotgh pull my-template --merge  # Merge only (no deletions)
```

### diff

Shows differences between a template and the current directory without applying changes.

**Flags:**
- `-r, --reverse`: Show differences for push direction (current → template)
- `--merge`: Show merge mode differences (no deletions)

**Exit codes:**
- 0: No differences found
- 1: Differences found or error occurred

### edit

Opens a template in the user's preferred editor. If the template doesn't exist:
- With `--create` flag: Creates the template directory
- Without flag: Prompts to create it

### update

Downloads from GitHub Releases with checksum validation; skips if running dev build.
