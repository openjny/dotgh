# User Guide

## Installation

### Quick Install (Recommended)

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

### Installation Options

#### Custom Installation Directory

```bash
# Linux / macOS
DOTGH_INSTALL_DIR=/custom/path curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

```powershell
# Windows
$env:DOTGH_INSTALL_DIR = "C:\tools\dotgh"; irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

#### Specific Version

```bash
# Linux / macOS
DOTGH_VERSION=v0.1.1 curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
```

```powershell
# Windows
$env:DOTGH_VERSION = "v0.1.1"; irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
```

### Alternative Installation Methods

#### Using Go

```bash
go install github.com/openjny/dotgh/cmd/dotgh@latest
```

#### Manual Download

Download the latest binary from [GitHub Releases](https://github.com/openjny/dotgh/releases/latest).

**Linux (amd64):**

```bash
curl -LO https://github.com/openjny/dotgh/releases/latest/download/dotgh_linux_amd64.tar.gz
tar xzf dotgh_linux_amd64.tar.gz
sudo mv dotgh /usr/local/bin/
```

**macOS (Apple Silicon):**

```bash
curl -LO https://github.com/openjny/dotgh/releases/latest/download/dotgh_darwin_arm64.tar.gz
tar xzf dotgh_darwin_arm64.tar.gz
sudo mv dotgh /usr/local/bin/
```

**Windows:**

Download `dotgh_windows_amd64.zip` from [Releases](https://github.com/openjny/dotgh/releases/latest), extract, and add to your PATH.

---

## Commands

### `dotgh list`

List all available templates.

```bash
dotgh list
```

### `dotgh pull <template>`

Pull a template to the current directory. This copies the `.github` directory from the template.

```bash
dotgh pull my-template

# Force overwrite existing files
dotgh pull my-template -f
```

### `dotgh push <template>`

Save the current directory's `.github` as a template.

```bash
dotgh push my-template

# Force overwrite existing template
dotgh push my-template -f
```

### `dotgh delete <template>`

Delete a template.

```bash
dotgh delete my-template

# Skip confirmation
dotgh delete my-template -f
```

### `dotgh version`

Display version information.

```bash
dotgh version
```

### `dotgh update`

Update dotgh to the latest version.

```bash
# Update to latest
dotgh update

# Check for updates without installing
dotgh update --check
```

### `dotgh config list`

Display the current configuration in YAML format.

```bash
dotgh config list
```

Example output:

```yaml
# Config file: ~/.config/dotgh/config.yaml
includes:
  - "AGENTS.md"
  - ".github/agents/*.agent.md"
  - ".github/copilot-chat-modes/*.chatmode.md"
  - ".github/copilot-instructions.md"
  - ".github/instructions/*.instructions.md"
  - ".github/prompts/*.prompt.md"
  - ".vscode/mcp.json"
```

### `dotgh config edit`

Open the configuration file in your preferred editor.

```bash
dotgh config edit
```

If the config file doesn't exist, it will be created with default values first.

---

## Configuration

### Config File Location

dotgh uses a YAML configuration file located at:

| Platform | Location |
|----------|----------|
| Linux/macOS | `~/.config/dotgh/config.yaml` |
| Windows | `%LOCALAPPDATA%\dotgh\config.yaml` |

### Customizing Target Patterns

You can customize which files are managed by templates by creating a `config.yaml`:

```yaml
editor: "code --wait"  # Optional: override the default editor
includes:
  - "AGENTS.md"
  - ".github/agents/*.agent.md"
  - ".github/copilot-chat-modes/*.chatmode.md"
  - ".github/copilot-instructions.md"
  - ".github/instructions/*.instructions.md"
  - ".github/prompts/*.prompt.md"
  - ".vscode/mcp.json"
```

### Editor Configuration

The `editor` field is optional and specifies the editor to use for `dotgh config edit`.

If not set, dotgh uses the following priority order:

1. `VISUAL` environment variable
2. `EDITOR` environment variable
3. `GIT_EDITOR` environment variable
4. Platform default (`vi` on Linux/macOS, `notepad` on Windows)

For GUI editors like VS Code or Sublime Text, the `--wait` flag is automatically added to ensure the command waits until the editor is closed.

#### Default Targets

If no config file exists, the following default patterns are used:

- `AGENTS.md` - AI agent instructions
- `.github/agents/*.agent.md` - Custom agent profiles
- `.github/copilot-chat-modes/*.chatmode.md` - Custom chat modes
- `.github/copilot-instructions.md` - GitHub Copilot instructions
- `.github/instructions/*.instructions.md` - Custom instruction files
- `.github/prompts/*.prompt.md` - Prompt templates
- `.vscode/mcp.json` - VS Code MCP server configuration

#### Glob Pattern Support

Target patterns support standard glob syntax:

- `*` matches any sequence of characters (except path separators)
- `?` matches any single character
- `[abc]` matches any character in the set

Examples:

```yaml
includes:
  - "*.md"                              # All markdown files in root
  - ".github/prompts/*.prompt.md"       # All prompt files
  - "config/*.json"                     # All JSON files in config/
```

---

## Template Storage

Templates are stored following the XDG Base Directory Specification:

| Platform | Location |
|----------|----------|
| Linux/macOS | `~/.config/dotgh/templates/` |
| Windows | `%LOCALAPPDATA%\dotgh\templates\` |

---

## Updating

You can update `dotgh` using the built-in update command:

```bash
dotgh update
```

Or re-run the installation script to get the latest version.
