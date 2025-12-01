#!/bin/bash
# Mock dotgh for demo recording

case "$1" in
  "list")
    echo "Available templates:"
    echo "  bicep-mslearn-mcp"
    echo "  node-playwright-mcp"
    echo "  python"
    echo ""
    echo "3 template(s) found"
    echo ""
    ;;
  "pull")
    echo "Pulling template '$2'..."
    echo "  ✓ AGENTS.md"
    echo "  ✓ .github/copilot-instructions.md"
    echo "  ✓ .github/instructions/bicep.instructions.md"
    echo "  ✓ .github/prompts/spec.prompt.md"
    echo "  ✓ .github/prompts/plan.prompt.md"
    echo "  ✓ .github/prompts/refactor.prompt.md"
    echo "  ✓ .vscode/mcp.json"
    echo ""
    echo "Template applied successfully!"
    echo ""
    ;;
  *)
    echo "dotgh - AI coding assistant config manager"
    ;;
esac
