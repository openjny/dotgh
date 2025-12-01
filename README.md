# dotgh

[![CI](https://github.com/openjny/dotgh/actions/workflows/ci.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/ci.yml)
[![Release](https://github.com/openjny/dotgh/actions/workflows/release.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/openjny/dotgh)](https://go.dev/)
[![License](https://img.shields.io/github/license/openjny/dotgh)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/openjny/dotgh)](https://github.com/openjny/dotgh/releases/latest)

A cross-platform CLI tool to manage `.github` directory for GitHub Copilot / VS Code users.

## Overview

`dotgh` allows you to easily apply, update, and manage AI coding guidelines and configuration templates across multiple projects.

## Features

- **Template Management**: Create, list, and delete templates
- **Apply**: Apply templates to your current project
- **Push**: Save your current project's configuration as a template
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Self-Update**: Easily update to the latest version

## Installation

### Using Go

```bash
go install github.com/openjny/dotgh/cmd/dotgh@latest
```

### Download from GitHub Releases

Download the latest binary from [GitHub Releases](https://github.com/openjny/dotgh/releases/latest).

#### Linux / macOS

```bash
# Linux (amd64)
curl -LO https://github.com/openjny/dotgh/releases/latest/download/dotgh_linux_amd64.tar.gz
tar xzf dotgh_linux_amd64.tar.gz
sudo mv dotgh /usr/local/bin/

# macOS (Apple Silicon)
curl -LO https://github.com/openjny/dotgh/releases/latest/download/dotgh_darwin_arm64.tar.gz
tar xzf dotgh_darwin_arm64.tar.gz
sudo mv dotgh /usr/local/bin/
```

#### Windows

Download `dotgh_windows_amd64.zip` from [Releases](https://github.com/openjny/dotgh/releases/latest), extract, and add to your PATH.

## Usage

```bash
dotgh list                      # List available templates
dotgh apply <template> [-f]     # Apply template to current directory
dotgh push <template> [-f]      # Save current directory as template
dotgh delete <template> [-f]    # Delete a template
dotgh version                   # Display version information
dotgh update                    # Update dotgh to latest version
dotgh update --check            # Check for updates without installing
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
