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
dotgh apply <template> [-f]     # Apply template to current directory
dotgh push <template> [-f]      # Save current directory as template
dotgh delete <template> [-f]    # Delete a template
dotgh version                   # Display version information
dotgh update                    # Update dotgh to latest version (planned)
```

## Template Storage

Templates are stored following the XDG Base Directory Specification:

- **Linux/macOS**: `~/.config/dotgh/templates/`
- **Windows**: `%LOCALAPPDATA%/dotgh/templates/`

## Development

```bash
make build          # Build the binary
make test           # Run tests
make lint           # Run linter
make release-check  # Validate goreleaser configuration
make release-dry    # Run goreleaser in snapshot mode
```

## Releases

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. To create a new release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This will automatically build binaries for all supported platforms and create a GitHub Release.
