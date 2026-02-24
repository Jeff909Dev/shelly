# Shelly AI

AI-powered terminal assistant that converts natural language to shell commands.

## Build & Run

```bash
go build -o q .        # Build binary
go build ./...         # Verify all packages compile
./q [request]          # Run with a query
./q config             # Open config TUI
```

## Project Structure

```
main.go            → Entry point (delegates to cli.RootCmd)
cli/cli.go         → Main TUI (Bubble Tea) - input, streaming, clipboard, history cmds
config/config.go   → YAML config loading/saving (~/.shelly-ai/config.yaml)
config/cli.go      → Config TUI menu (model selection, theme selection)
config/config.yaml → Embedded default config
llm/llm.go         → LLM HTTP client with SSE streaming
llm/provider.go    → Provider interface + detection (OpenAI, Anthropic, Gemini, Plugin)
llm/provider_*.go  → Provider implementations
types/types.go     → Shared types (ModelConfig, Message, Payload, Preferences)
util/util.go       → Terminal width, code extraction, browser open
theme/             → Theme system (6 built-in themes)
history/           → Conversation history (JSONL storage)
plugin/            → Plugin system (JSON-RPC 2.0 provider plugins)
examples/          → Example plugin implementation
```

## Key Architecture Decisions

- **Go module name is `q`** (not `shelly-ai`) — all imports use `q/cli`, `q/config`, etc.
- **Bubble Tea** (Elm architecture) for all TUI — state machine: Loading → ReceivingInput → ReceivingResponse
- **Streaming via SSE** — `llm.processStream()` uses `strings.Builder` for O(n) concatenation
- **Config stored at** `~/.shelly-ai/config.yaml` with automatic backup at `~/.shelly-ai/.backup-config.yaml`
- **Binary name** is `shelly-ai` with symlink `q` for quick access
- **Auth via env vars** — config stores env var *names* (e.g. `OPENAI_API_KEY`), not secrets

## Code Style

- No panics in production paths — use graceful error returns
- Errors in TUI rendering fall back to raw text display
- Keep `types/` as a pure data package with no dependencies
- Use `strings.Builder` for string concatenation in loops (never `+=`)

## Testing

```bash
go test ./...          # Run all tests before commits
go build ./...         # Verify compilation
```

- Use **table-driven tests** as the standard pattern
- Test files: `llm/llm_test.go`, `llm/provider_test.go`, `util/util_test.go`, `config/config_test.go`
- Use `httptest.NewServer` for testing streaming/HTTP behavior
- Provider tests should cover `ParseStreamLine`, `BuildRequestBody`, and `SetHeaders`

## Common Pitfalls

- `util.GetTermSafeMaxWidth()` has a minimum of 20 to prevent negative widths
- `config.writeConfigToFile()` always saves a backup — double I/O is intentional
- The `sahilm/fuzzy` dependency appears unused but is pulled by bubbles — do not remove manually
