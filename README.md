# dotgh

[![CI](https://github.com/openjny/dotgh/actions/workflows/ci.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/ci.yml)
[![Release](https://github.com/openjny/dotgh/actions/workflows/release.yml/badge.svg)](https://github.com/openjny/dotgh/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/v/release/openjny/dotgh)](https://github.com/openjny/dotgh/releases/latest)

A CLI tool to manage AI coding assistant configuration templates.

<p align="center">
  <img src="assets/demo.svg" alt="dotgh demo" width="600">
</p>

## 💡 Why dotgh?

If you're using AI coding assistants like GitHub Copilot or Cursor, you've probably noticed yourself creating similar config files over and over again — `.github/copilot-instructions.md`, `.github/prompts/my.prompt.md`, `AGENTS.md`, and so on...

`dotgh` is a cross-platform tool that lets you save and apply these config files as templates. When starting a new project, just run `dotgh pull my-template` and you're good to go 👌.

## 📁 What it manages

By default, `dotgh` manages these files as template components:

- `AGENTS.md`
- `.github/agents/*.agent.md`
- `.github/copilot-chat-modes/*.chatmode.md`
- `.github/copilot-instructions.md`
- `.github/instructions/*.instructions.md`
- `.github/prompts/*.prompt.md`
- `.vscode/mcp.json`

This is customizable via `~/.config/dotgh/config.yaml`. See [User Guide](docs/user-guide.md#configuration) for details.

## 📦 Install

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

See [User Guide](docs/user-guide.md#installation) for more installation options.

## 🚀 Usage

```bash
dotgh list                  # List templates
dotgh pull <template>       # Get a template
dotgh push <template>       # Save as a template
dotgh delete <template>     # Delete a template
dotgh config show           # Show current configuration
dotgh config edit           # Edit configuration file
dotgh update                # Update dotgh to latest version
```

See [User Guide](docs/user-guide.md#commands) for detailed command usage.

## 📖 Documentation

- [User Guide](docs/user-guide.md) - Installation, commands, configuration
- [Development](docs/development.md) - Contributing, testing, releases

## ⚖️ License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
