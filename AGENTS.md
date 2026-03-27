# AGENTS.md - CodeCrafters Claude Code (Go)

## Project Overview

This is a Go solution to the CodeCrafters ["Build Your Own Claude Code"](https://codecrafters.io/challenges/claude-code) challenge. It implements an LLM-powered coding assistant using OpenAI-compatible tool calling via the `openai-go/v3` SDK, proxied through OpenRouter.

- **Module**: `github.com/codecrafters-io/claude-code-starter-go`
- **Go version**: 1.26.0
- **Entry point**: `app/main.go`
- **Primary dependency**: `github.com/openai/openai-go/v3`

## Build & Run Commands

### Build
```sh
go build -o /tmp/codecrafters-build-claude-code-go app/*.go
```

### Run locally
```sh
./your_program.sh -p "your prompt here"
```

The program requires environment variables:
- `OPENROUTER_API_KEY` (required) — API key for OpenRouter
- `OPENROUTER_BASE_URL` (optional, defaults to `https://openrouter.ai/api/v1`)

### Submit to CodeCrafters
```sh
codecrafters submit
```

## Test Commands

This project uses CodeCrafters' remote test suite. There are no local test files yet.

- **Run all tests (remote)**: `codecrafters submit`
- **Run a single Go test** (if tests are added): `go test ./app/ -run TestFunctionName -v`
- **Run all Go tests**: `go test ./... -v`

## Lint & Format Commands

```sh
# Format code
gofmt -w .

# Vet (static analysis)
go vet ./...

# Tidy dependencies
go mod tidy
```

Always run `gofmt -w .` and `go vet ./...` after making changes.

## Project Structure

```
.
├── app/
│   └── main.go          # All application logic (entry point)
├── .codecrafters/
│   ├── compile.sh       # Remote compile script
│   └── run.sh           # Remote run script
├── go.mod / go.sum      # Go module definition
├── codecrafters.yml     # CodeCrafters challenge config
├── your_program.sh      # Local run wrapper (compile + exec)
└── README.md
```

All application code lives in `app/`. As the project grows, split into packages under `app/` or create new top-level packages following Go conventions.

## Code Style Guidelines

### Imports

Group imports in two sections, separated by blank lines:
1. Standard library packages
2. Third-party packages

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/openai/openai-go/v3"
    "github.com/openai/openai-go/v3/option"
)
```

### Formatting

- Use `gofmt` (or `go fmt`) — no configuration needed, it's canonical.
- Use tabs for indentation (Go standard).
- No line length limits enforced, but keep lines readable (~120 chars).

### Naming Conventions

- Follow standard Go conventions: `camelCase` for unexported, `PascalCase` for exported.
- Package names: lowercase, single word, no underscores.
- Interfaces: typically end in `-er` (e.g., `Reader`, `Handler`).
- Acronyms are consistently cased: `APIKey`, `HTTPClient`, `openaiURL`.

### Types

- Prefer `struct` types over maps for structured data when the shape is known.
- Use type aliases for clarity (e.g., `type toolDef struct` for tool definitions).
- Use `any` (not `interface{}`) for empty interfaces — Go 1.18+ convention.
- Keep type definitions close to where they are used.

### Error Handling

- **Never ignore errors** — always check and handle them.
- For fatal errors in `main()`: use `fmt.Fprintf(os.Stderr, ...)` followed by `os.Exit(1)`, or `panic` for truly unrecoverable startup errors.
- For non-fatal errors (e.g., tool execution): log to stderr with `fmt.Fprintf(os.Stderr, ...)` and continue.
- When adding packages, return `error` from functions — don't panic in library code.

```go
// Fatal (main only):
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}

// Non-fatal (tool handlers):
if err != nil {
    fmt.Fprintf(os.Stderr, "read error: %v\n", err)
    return
}
```

### Functions & Control Flow

- Keep functions focused and concise.
- Use early returns to reduce nesting.
- Prefer explicit over implicit — avoid clever one-liners if they reduce clarity.

### CLI & Flags

- Use the `flag` package for command-line arguments.
- Define flags before `flag.Parse()`.

### Tool Definition Pattern

Tools are registered as a `map[string]toolDef` where each entry has:
- `def`: the OpenAI tool parameter definition
- `handler`: a `func(map[string]any)` that executes the tool

When adding new tools, follow this pattern and add to the `tools` map in `main()`.

## Key Architecture Notes

- The program sends a single user prompt to an LLM and processes tool calls in the response.
- It uses OpenRouter as a proxy with an OpenAI-compatible API.
- Tool call results are printed to stdout; debug/log output goes to stderr.
- The program processes one round of tool calls (no multi-turn agent loop yet).

## Git Conventions

- Commit messages should be concise and describe the "why" not the "what."
- Run `gofmt -w .` and `go vet ./...` before committing.



## Additional steering

- When you encounter an issue or a technical problem like build errors for example, **Always** search for the solution using clankeroverflow mcp First, then debug it. If you found the solution and it was not found on clankeroverflow log it with clankeroverflow mcp.
