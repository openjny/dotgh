# dotgh Design

## Architecture

Implemented as a single-binary CLI application in Go.

### Tech Stack

- **Language**: Go
- **CLI Framework**: `spf13/cobra`
- **Self-Update**: `minio/selfupdate` or `rhysd/go-github-selfupdate` (planned)
- **Release**: `goreleaser` (integrated with GitHub Actions)

## Directory Structure

Complies with the XDG Base Directory Specification.

### Configuration & Template Storage Location

- **Linux/macOS**: `~/.config/dotgh/templates/`
- **Windows**: `~/AppData/Local/dotgh/templates/` (or unified to `~/.config/dotgh/templates/`)

### Repository Structure

```
dotgh/
├── .devcontainer/
│   └── devcontainer.json # Development container configuration
├── cmd/
│   └── dotgh/
│       └── main.go       # Entry point
├── internal/
│   ├── commands/         # Implementation of each subcommand (list, apply, push, delete)
│   ├── config/           # Configuration loading
│   ├── fileutil/         # File operation utilities (Copy, ExistCheck)
│   └── updater/          # Self-update logic
├── docs/                 # Documentation
├── install.sh            # Linux/macOS installer
├── install.ps1           # Windows installer
├── go.mod
└── README.md
```

## CLI Interface

```bash
dotgh [command] [flags]
```

### Command List

| Command   | Arguments    | Options       | Description                                         | Status      |
| --------- | ------------ | ------------- | --------------------------------------------------- | ----------- |
| `list`    | None         | None          | Display a list of available templates               | Implemented |
| `apply`   | `<template>` | `-f, --force` | Apply a template to the current directory           | Implemented |
| `push`    | `<template>` | `-f, --force` | Save the current directory's settings as a template | Implemented |
| `delete`  | `<template>` | `-f, --force` | Delete a template                                   | Planned     |
| `update`  | None         | None          | Update dotgh itself to the latest version           | Planned     |
| `version` | None         | None          | Display version information                         | Planned     |

## Data Structures

### Default Targets

By default, the following files/directories are treated as template components:

- `.github/` (directory)
- `.vscode/` (directory)
- `AGENTS.md` (file)

※ In the future, these will be customizable via `dotgh.yaml`.

## Process Flow

### Apply Flow

1. Check for the existence of the template directory
2. Scan the target files (`.github/`, etc.) within the template
3. Copy to the current directory
   - If existing files are present:
     - With `-f` specified: Overwrite
     - Without `-f` specified: Skip (display message)
   - New files: Create

### Push Flow

1. Scan the target files (`.github/`, etc.) in the current directory
2. Copy to the template directory
   - If the template does not exist: Create new
   - If the template exists:
     - With `-f` specified: Overwrite
     - Without `-f` specified: Skip (display message)
