# dotgh

A cross-platform CLI tool to manage `.github` directory for GitHub Copilot / VS Code users.

## Overview

`dotgh` allows you to easily apply, update, and manage AI coding guidelines and configuration templates across multiple projects.

## Features

- **Template Management**: Create, list, and delete templates
- **Apply**: Apply templates to your current project
- **Push**: Save your current project's configuration as a template
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Self-Update**: Easily update to the latest version

## Installation (Planned)

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

## Usage

```bash
dotgh list                      # List available templates
dotgh apply <template> [-f]     # Apply template to current directory (planned)
dotgh push <template> [-f]      # Save current directory as template (planned)
dotgh delete <template>         # Delete a template (planned)
dotgh update                    # Update dotgh to latest version (planned)
```

## Template Storage

Templates are stored following the XDG Base Directory Specification:

- **Linux/macOS**: `~/.config/dotgh/templates/`
- **Windows**: `%LOCALAPPDATA%/dotgh/templates/`

## Development

```bash
make build    # Build the binary
make test     # Run tests
make lint     # Run linter
```
