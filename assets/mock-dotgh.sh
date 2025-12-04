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
    echo "Pulling template '$2' (full sync):"
    echo "  + AGENTS.md"
    echo "  + .github/instructions/bicep.instructions.md"
    echo "  + .github/prompts/refactor.prompt.md"
    echo "  + .github/agents/plan.agent.md"
    echo "  + .github/agents/implement.agent.md"
    echo "  + .vscode/mcp.json"
    echo ""
    echo "Done: 6 added"
    echo ""
    ;;
  "diff")
    echo "Comparing with template '$2':"
    echo ""
    echo "Local files are in sync with template."
    echo ""
    ;;
  *)
    echo "dotgh - AI coding assistant config manager"
    ;;
esac
