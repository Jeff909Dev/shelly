<h1 align="center">Shelly AI</h1>
<p align="center">AI-powered shell assistant. Natural language to commands.</p>

<p align="center">
<a href="https://github.com/Jeff909Dev/shelly-ai/releases"><img src="https://img.shields.io/github/v/release/Jeff909Dev/shelly-ai?style=flat-square" alt="Release"></a>
<a href="https://github.com/Jeff909Dev/shelly-ai/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Jeff909Dev/shelly-ai?style=flat-square" alt="License"></a>
<img src="https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square" alt="Go">
</p>

<br>

```
$ q find all go files modified in the last hour
find . -name "*.go" -mmin -60
```

## Features

- **Natural language to shell commands** — describe what you want, get the command
- **Multi-provider** — OpenAI, Anthropic Claude, Google Gemini, Groq, local models
- **Pipe support** — `cat error.log | q "what went wrong"`
- **Themes** — 6 built-in color schemes (default, dracula, catppuccin, nord, tokyo-night, gruvbox)
- **Conversation history** — search and revisit past queries
- **Plugin system** — extend with custom providers via JSON-RPC
- **Auto-copy** — press Enter to copy the command to clipboard
- **Follow-up** — refine responses in the same session
- **Streaming** — real-time response rendering with syntax highlighting

## Install

### Homebrew

```bash
brew tap Jeff909Dev/tap
brew install shelly-ai
```

### Linux

```bash
curl https://raw.githubusercontent.com/Jeff909Dev/shelly-ai/main/install.sh | bash
```

### From source

```bash
go install github.com/Jeff909Dev/shelly-ai@latest
```

## Quick Start

```bash
# Set your API key
export OPENAI_API_KEY=sk-...

# Ask anything
q list files larger than 100MB
q "how do I undo the last git commit"
q create a tarball of the src directory

# Pipe input
cat error.log | q "explain this error"
git diff | q "summarize these changes"
echo '{"name":"test"}' | q "convert to yaml"
```

## Providers

| Provider | Models | Auth Env Var |
|----------|--------|-------------|
| OpenAI | gpt-4.1, gpt-4.1-mini | `OPENAI_API_KEY` |
| Anthropic | claude-sonnet-4-5, claude-opus-4-6 | `ANTHROPIC_API_KEY` |
| Google | gemini-2.5-flash, gemini-2.5-pro | `GEMINI_API_KEY` |
| Groq | llama-3.3-70b | `GROQ_API_KEY` |
| Local | Any OpenAI-compatible endpoint | — |

Switch models:

```bash
q config  # → Change Default Model
```

Or set in `~/.shelly-ai/config.yaml`.

## Themes

Six built-in themes. Switch via `q config` → Change Theme.

Available: `default` · `dracula` · `catppuccin-mocha` · `nord` · `tokyo-night` · `gruvbox`

## History

Conversations are saved automatically when `history_enabled: true` (default).

```bash
q history                    # List recent conversations
q history search "docker"    # Search by content
q history show abc123        # View full conversation
q history clear              # Clear all history
```

## Plugins

Extend Shelly AI with custom providers. Plugins are executables that communicate via JSON-RPC 2.0 over stdin/stdout.

### Creating a plugin

1. Create a directory in `~/.shelly-ai/plugins/your-plugin/`
2. Add a `plugin.yaml`:

```yaml
name: your-plugin
type: provider
executable: your-binary
```

3. Build your executable implementing three RPC methods: `set_headers`, `build_request_body`, `parse_stream_line`
4. Reference it in config:

```yaml
models:
  - name: my-custom-model
    endpoint: https://your-api.com/v1/chat
    auth_env_var: YOUR_API_KEY
    plugin: your-plugin
```

See `examples/example-provider-plugin/` for a complete reference implementation.

## Configuration

Config lives at `~/.shelly-ai/config.yaml`. Edit via `q config` or directly.

```yaml
preferences:
  default_model: gpt-4.1
  theme: default
  history_enabled: true
  history_max_days: 30

models:
  - name: gpt-4.1
    endpoint: https://api.openai.com/v1/chat/completions
    auth_env_var: OPENAI_API_KEY
    prompt:
      - role: system
        content: "You are a terminal assistant..."
```

### Local models

Any OpenAI-compatible server works (Ollama, llama.cpp, vLLM):

```yaml
models:
  - name: local-llama
    endpoint: http://localhost:11434/v1/chat/completions
    auth_env_var: OPENAI_API_KEY
```

### Recovery

```bash
q config revert   # Restore from automatic backup
q config reset    # Reset to defaults
```

## Contributing

Contributions welcome! See [GitHub](https://github.com/Jeff909Dev/shelly-ai).

## License

MIT
