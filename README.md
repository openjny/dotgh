# dotgh

[![CI](https://github.com/openjny/dotgh/actions/workflows/ci.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/ci.yml)
[![Release](https://github.com/openjny/dotgh/actions/workflows/release.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/openjny/dotgh)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/openjny/dotgh)](https://github.com/openjny/dotgh/releases/latest)

A CLI tool to manage AI coding assistant configuration templates.

<p align="center">
  <img src="assets/demo.gif" alt="dotgh demo" width="600">
</p>

## 💡 Why dotgh?

If you're using AI coding assistants like GitHub Copilot or Cursor, you've probably noticed yourself creating similar config files over and over again — `copilot-instructions.md`, `.github/prompts/myprompts.md`, `AGENTS.md`, and so on...

`dotgh` is a cross-platform tool that lets you save and apply these config files as templates. When starting a new project, just run `dotgh pull my-awesome-template` and you're good to go 👌.

## 📁 What it manages

By default, `dotgh` manages these files:

- `AGENTS.md` - AI agent instructions
- `.github/copilot-instructions.md` - GitHub Copilot instructions
- `.github/instructions/*.instructions.md` - Custom instruction files
- `.github/prompts/*.prompt.md` - Prompt templates
- `.vscode/mcp.json` - VS Code MCP server configuration

> Customizable via `~/.config/dotgh/config.yaml`. See [User Guide](docs/user-guide.md) for details.

## 📦 Install

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

> See [User Guide](docs/user-guide.md) for more installation options.

## 🚀 Usage

```bash
dotgh list                  # List templates
dotgh pull <template>       # Get a template
dotgh push <template>       # Save as a template
dotgh delete <template>     # Delete a template
dotgh update                # Update dotgh to latest version
```

## 📖 Documentation

- [User Guide](docs/user-guide.md) - Installation, commands, configuration
- [Development](docs/development.md) - Contributing, testing, releases
