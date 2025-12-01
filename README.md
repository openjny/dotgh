# dotgh

[![CI](https://github.com/openjny/dotgh/actions/workflows/ci.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/ci.yml)
[![Release](https://github.com/openjny/dotgh/actions/workflows/release.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/openjny/dotgh)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/openjny/dotgh)](https://github.com/openjny/dotgh/releases/latest)

A CLI tool to manage `.github` directory templates for GitHub Copilot / VS Code users.

## Install

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

> See [User Guide](docs/user-guide.md) for more installation options.

## Usage

```bash
dotgh list                  # List templates
dotgh apply <template>      # Apply template to current directory
dotgh push <template>       # Save .github as template
dotgh delete <template>     # Delete template
dotgh update                # Update to latest version
```

## Documentation

- [User Guide](docs/user-guide.md) - Installation, commands, configuration
- [Development](docs/development.md) - Contributing, testing, releases
