# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-based educational guide for learning Claude API and ecosystem (14 chapters, 00-13). Two learning paths: CLI-only (no API key) and full API path.

## Build & Run

```bash
# Run individual example
go run 01-api-basics/01_hello_claude.go

# Run project examples
cd 10-projects/chatbot && go run main.go

# Build MCP server
cd 07-mcp && go build -o weather-server 01_mcp_server.go

# Verify all non-ignore files compile
go build ./...

# Generate PDF manual (requires venv with weasyprint)
source /tmp/pdf-venv/bin/activate
DYLD_FALLBACK_LIBRARY_PATH="/opt/homebrew/lib" python3 generate_pdf.py
```

## Code Conventions

- **Every standalone example file** uses `package main` and `//go:build ignore` (for CLI/exec-based examples in chapters 00, 11, 12) or no build tag (for API-based examples in chapters 01-10)
- **All comments and output** are in Chinese (中文)
- **Error handling**: `log.Fatalf()` with Chinese messages for fatal errors
- **Model constant**: `anthropic.ModelClaudeSonnet4_5_20250929` for API examples; `--model sonnet` for CLI examples
- **CLI examples** call `claude` via `os/exec` and parse output with `--output-format json`
- **No shared internal packages**: each `.go` file is a self-contained, independently runnable program

## Architecture

```
00-cli-quick-start/   CLI basics (//go:build ignore, os/exec)
01-api-basics/        SDK init, Messages API, streaming
02-prompt-engineering/ Few-shot, chain-of-thought, structured output
03-tool-use/          Function calling, tool definitions, agent loops
04-vision/            Image analysis
05-extended-thinking/ Deep reasoning, thinking budget
06-advanced-patterns/ Caching, batching, cost control
07-mcp/               MCP server implementation (mcp-go SDK)
08-agent-patterns/    ReAct, multi-agent, memory
09-claude-code/       Claude Code CLI guide (markdown only)
10-projects/          3 complete apps: chatbot, code-reviewer, rag-assistant
11-cli-mastery/       Advanced CLI flags (//go:build ignore, os/exec)
12-skills/            SKILL.md authoring + Go test harness
13-hooks-system/      Hooks automation: security guards, quality checks, context injection
```

- Chapters 00, 09, 11, 12, 13 do **not** require `ANTHROPIC_API_KEY`
- Chapter 10 projects use struct-based organization (`ChatBot`, `CodeReviewer` types)
- Chapter 12 `skills/` contains 4 example SKILL.md files with YAML frontmatter

## Key Dependencies

- `github.com/anthropics/anthropic-sdk-go` — Claude API client
- `github.com/mark3labs/mcp-go` — MCP server framework (chapter 07)

## Adding New Examples

1. Place in the appropriate chapter directory with sequential numbering (`NN_descriptive_name.go`)
2. Add `//go:build ignore` if the file uses `os/exec` to call `claude` CLI (not the SDK)
3. Include a file header comment with: chapter/example number, what it demonstrates, how to run it
4. Update the chapter's `README.md` file listing
5. If adding a new chapter, update root `README.md` learning roadmap table and `generate_pdf.py` CHAPTERS list
