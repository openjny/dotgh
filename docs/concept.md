# Concept

## Background

In development projects, managing configuration files and prompt files (such as `.github/`, `AGENTS.md`, `.vscode/`, etc.) for effectively utilizing AI coding assistants (GitHub Copilot, Claude, ChatGPT, etc.) has become important.
These files are often manually copied and pasted for each project, making management cumbersome. Additionally, there is a need to apply unified "instructions for AI (guidelines)" across teams or individuals.

## Goal

To provide a cross-platform CLI tool that allows applying, updating, and deleting AI coding guidelines and configuration file templates with a single command.
This will shorten the setup time at the start of a project and make it easy to deploy best practices for AI utilization.

## Scope

- **Target Users**: Developers using AI coding tools
- **Target Platforms**: Windows, Linux, macOS (Cross-platform)
- **Key Features**:
  - List templates
  - Apply templates (Apply)
  - Update templates (Push / Update)
  - Delete templates (Delete)
  - Self-update function

## Requirements

### Functional Requirements

1. **Template Management**
   - Ability to manage user-defined template directories.
   - Default targets for templates should be `.github/`, `.vscode/`, and `AGENTS.md`.
2. **CLI Operations**
   - `list`: Display available templates.
   - `apply`: Expand the specified template into the current directory (option to skip/overwrite existing files).
   - `push`: Save/update the current directory's settings as a template.
   - `delete`: Delete a template.
3. **Self-Update**
   - Ability to fetch and update to the latest binary from GitHub Releases using the `update` command.

### Non-Functional Requirements

- **Single Binary**: Must operate without dependencies.
- **Configuration Flexibility**: Future capability to customize monitored files via a configuration file (`.dotgh.yaml`).
- **Common CLI Conventions**: Follow common CLI tool conventions for usability (e.g., adhere to XDG Base Directory Specification like `$XDG_CONFIG_HOME/dotgh` on Linux/macOS, and `%APPDATA%\dotgh` on Windows).
