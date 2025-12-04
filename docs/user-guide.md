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

Pull a template to the current directory with Git-style sync behavior.

By default, performs a **full sync**:
- Adds new files from the template
- Updates modified files
- Deletes files that exist locally but not in the template

```bash
# Full sync with confirmation prompt
dotgh pull my-template

# Full sync without confirmation
dotgh pull my-template --yes

# Merge mode: only add/update, no deletions
dotgh pull my-template --merge
```

**Options:**
- `-m, --merge`: Only add and update files, don't delete local-only files
- `-y, --yes`: Skip the confirmation prompt

### `dotgh push <template>`

Save the current directory's settings as a template with Git-style sync behavior.

By default, performs a **full sync**:
- Adds new files to the template
- Updates modified files in the template
- Deletes files in the template that don't exist locally

If the template doesn't exist, it will be created.

```bash
# Full sync with confirmation prompt
dotgh push my-template

# Full sync without confirmation
dotgh push my-template --yes

# Merge mode: only add/update, no deletions
dotgh push my-template --merge
```

**Options:**
- `-m, --merge`: Only add and update files, don't delete files from the template
- `-y, --yes`: Skip the confirmation prompt

### `dotgh diff <template>`

Show differences between a template and the current directory without applying changes.

```bash
# Show what pull would do (template → current)
dotgh diff my-template

# Show what push would do (current → template)
dotgh diff my-template --reverse

# Show merge mode differences (no deletions)
dotgh diff my-template --merge
```

**Options:**
- `-r, --reverse`: Show differences for push direction (current → template)
- `--merge`: Show merge mode differences (no deletions)

**Output symbols:**
- `+ file`: File will be added
- `M file`: File will be modified
- `- file`: File will be deleted

### `dotgh delete <template>`

Delete a template.

```bash
dotgh delete my-template

# Skip confirmation
dotgh delete my-template -f
```

### `dotgh edit [template]`

Open a template or the templates directory in your preferred editor.

```bash
# Open templates directory
dotgh edit

# Open a specific template
dotgh edit my-template

# Create and open a new template
dotgh edit new-template --create
```

When called without arguments, opens the templates directory (`~/.config/dotgh/templates/`) in the editor. When a template name is provided, opens that specific template directory.

If the template doesn't exist:
- With `--create` flag: Creates the template directory automatically
- Without flag: Prompts you to create it

**Options:**
- `-c, --create`: Create the template if it doesn't exist

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

### `dotgh config show`

Display the current configuration in YAML format.

```bash
dotgh config show
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

`dotgh` uses a YAML configuration file located at:

| Platform | Location |
|----------|----------|
| Linux/macOS | `~/.config/dotgh/config.yaml` |
| Windows | `%LOCALAPPDATA%\dotgh\config.yaml` |

If the config file does not exist, default settings are used. You can create or edit the config file using the `dotgh config edit` command.

The configuration file supports the following fields:

```yaml
editor: "code --wait"
templates_dir: "~/my-templates"
includes:
  - "AGENTS.md"
  - ".github/agents/*.agent.md"
  - ".github/copilot-chat-modes/*.chatmode.md"
  - ".github/copilot-instructions.md"
  - ".github/instructions/*.instructions.md"
  - ".github/prompts/*.prompt.md"
  - ".vscode/mcp.json"
excludes:
  - ".github/prompts/local.prompt.md"
  - ".github/prompts/secret-*.prompt.md"
```

### editor:

The `editor` field is optional and specifies the editor to use for `dotgh edit` and `dotgh config edit` commands.

If not set, dotgh uses the following priority order:

1. `VISUAL` environment variable
2. `EDITOR` environment variable
3. `GIT_EDITOR` environment variable
4. Platform default (`vi` on Linux/macOS, `notepad` on Windows)

For GUI editors like VS Code or Sublime Text, the `--wait` flag is automatically added to ensure the command waits until the editor is closed.

### templates_dir

The `templates_dir` field is optional and specifies a custom location for the templates directory. This allows you to store your templates in a location other than the default.

**Default locations:**
- Linux/macOS: `~/.config/dotgh/templates/`
- Windows: `%LOCALAPPDATA%\dotgh\templates\`

**Features:**
- Supports tilde expansion (e.g., `~/my-templates`)
- Can be an absolute or relative path

**Example:**

```yaml
templates_dir: "~/dotfiles/dotgh-templates"
```

This is useful when:
- You want to keep templates in a version-controlled dotfiles repository
- You want to share templates across multiple machines via cloud storage
- You need to organize templates in a specific location for your workflow


### includes:

The `includes` field is required and specifies the files and directories to manage as template components. It supports glob patterns for flexible matching.

If no config file exists, the following default patterns are used:

- `AGENTS.md`
- `.github/agents/*.agent.md`
- `.github/copilot-chat-modes/*.chatmode.md`
- `.github/copilot-instructions.md`
- `.github/instructions/*.instructions.md`
- `.github/prompts/*.prompt.md`
- `.vscode/mcp.json`

It supports standard glob syntax:

- `*` matches any sequence of characters (except path separators)
- `?` matches any single character
- `[abc]` matches any character in the set

> **Note:** Recursive patterns (`**`) are not supported. Use explicit directory paths like `.github/prompts/*.prompt.md` instead of `**/*.prompt.md`.

#### Examples

**Claude Code:**

```yaml
includes:
  - "AGENTS.md"
  - "CLAUDE.md"
  - ".claude/settings.json"
excludes:
  - "CLAUDE.local.md"
  - ".claude/settings.local.json"
```

**Gemini CLI:**

```yaml
includes:
  - "AGENTS.md"
  - "GEMINI.md"
  - ".gemini/settings.json"
  - ".gemini/system.md"
```

**Cursor:**

```yaml
includes:
  - "AGENTS.md"
  - ".cursorrules"
  - ".cursor/rules/*.mdc"
```

**Windsurf:**

```yaml
includes:
  - "AGENTS.md"
  - ".windsurfrules"
  - ".windsurf/rules/*.md"
```

**Cline:**

```yaml
includes:
  - "AGENTS.md"
  - ".clinerules"
```

**Kilo Code:**

```yaml
includes:
  - "AGENTS.md"
  - ".kilocode/rules/*.md"
```

**Roo Code:**

```yaml
includes:
  - "AGENTS.md"
  - ".roorules"
  - ".roo/rules/*.mdc"
```

### excludes

The `excludes` field allows you to exclude specific files from template management, even if they match an `includes` pattern.

**Use cases:**

- Exclude project-specific or local configuration files
- Exclude sensitive files that shouldn't be shared across projects
- Fine-tune which files are included when using broad patterns

**Behavior:**

- Files matching any `excludes` pattern are filtered out after `includes` expansion
- `excludes` takes priority: if a file matches both, it is excluded
- Default: empty list (no exclusions)

**Example:**

```yaml
includes:
  - ".github/prompts/*.prompt.md"  # Include all prompt files
excludes:
  - ".github/prompts/local.prompt.md"      # Exclude specific file
  - ".github/prompts/secret-*.prompt.md"   # Exclude files matching pattern
```

---

## Template Storage

Templates are stored following the XDG Base Directory Specification:

| Platform | Location |
|----------|----------|
| Linux/macOS | `~/.config/dotgh/templates/` |
| Windows | `%LOCALAPPDATA%\dotgh\templates\` |

You can customize the templates directory location by setting `templates_dir` in your configuration file. See the [templates_dir](#templates_dir) section for details.

---

## Syncing Configuration Across Machines

The `sync` command allows you to synchronize your dotgh configuration and templates across multiple machines using a Git repository.

### Setting Up Sync

First, create a Git repository to store your sync data (e.g., on GitHub):

```bash
# Create a new repository for sync (e.g., github.com/username/dotgh-sync)
# Then initialize sync locally:
dotgh sync init git@github.com:username/dotgh-sync.git
```

### Sync Commands

#### `dotgh sync init <repository>`

Initialize sync with a Git repository.

```bash
# Using SSH
dotgh sync init git@github.com:user/dotgh-sync.git

# Using HTTPS
dotgh sync init https://github.com/user/dotgh-sync.git

# Specify a branch
dotgh sync init git@github.com:user/dotgh-sync.git --branch main
```

**Options:**
- `-b, --branch`: Branch to use for sync (default: `main`)

#### `dotgh sync push`

Push local configuration and templates to the remote repository.

```bash
# Push with auto-generated commit message
dotgh sync push

# Push with custom commit message
dotgh sync push -m "Update templates for new project"
```

**Options:**
- `-m, --message`: Custom commit message

#### `dotgh sync pull`

Pull configuration and templates from the remote repository.

```bash
dotgh sync pull
```

This will:
1. Pull the latest changes from the remote
2. Copy `config.yaml` to your local config directory
3. Copy templates to your local templates directory

#### `dotgh sync status`

Show the current sync status.

```bash
dotgh sync status
```

Example output:

```
Sync Status:
  Repository: git@github.com:user/dotgh-sync.git
  Branch: main
  Status: clean
  Sync directory: ~/.config/dotgh/.sync
```

### Typical Workflow

**On your primary machine:**

```bash
# 1. Initialize sync (one time)
dotgh sync init git@github.com:user/dotgh-sync.git

# 2. Push your config and templates
dotgh sync push -m "Initial sync"
```

**On a new machine:**

```bash
# 1. Install dotgh
curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash

# 2. Initialize sync with the same repository
dotgh sync init git@github.com:user/dotgh-sync.git

# 3. Pull your config and templates
dotgh sync pull
```

**Keeping machines in sync:**

```bash
# After making changes on any machine, push them:
dotgh sync push -m "Updated templates"

# On other machines, pull the changes:
dotgh sync pull
```

### Sync Configuration

The following configuration options are **planned for future releases** and are not yet implemented:

```yaml
# sync: Configuration for syncing settings across machines (PLANNED)
# These options are not yet functional - use command-line arguments instead.
# sync:
#   repo: "git@github.com:username/dotgh-sync.git"  # Default repository
#   branch: "main"                                   # Default branch
#   auto_commit: true                                # Auto-commit on push
```

Currently, use command-line arguments to specify the repository and branch:

```bash
dotgh sync init git@github.com:user/dotgh-sync.git --branch main
```

### Security Considerations

> ⚠️ **Warning:** Be careful about what you sync. The sync feature will upload your configuration and templates to a Git repository.

**Do NOT include in your templates:**
- API keys or tokens
- Passwords or secrets
- Private SSH keys
- Personal access tokens
- Any sensitive credentials

**Recommendations:**
- Use the `excludes` pattern to exclude sensitive files
- Use private repositories for sync
- Review your templates before pushing

---

## Updating

You can update `dotgh` using the built-in update command:

```bash
dotgh update
```

Or re-run the installation script to get the latest version.
