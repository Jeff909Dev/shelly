package llm

import (
	"encoding/json"
	. "q/types"
	"testing"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantType string
	}{
		{"openai", "https://api.openai.com/v1/chat/completions", "*llm.OpenAIProvider"},
		{"anthropic", "https://api.anthropic.com/v1/messages", "*llm.AnthropicProvider"},
		{"azure", "https://myresource.openai.azure.com/openai/deployments/gpt-4/chat/completions", "*llm.OpenAIProvider"},
		{"groq", "https://api.groq.com/openai/v1/chat/completions", "*llm.OpenAIProvider"},
		{"local", "http://localhost:11434/v1/chat/completions", "*llm.OpenAIProvider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := detectProvider(tt.endpoint)
			got := typeString(p)
			if got != tt.wantType {
				t.Errorf("detectProvider(%q) = %s, want %s", tt.endpoint, got, tt.wantType)
			}
		})
	}
}

func typeString(p Provider) string {
	return func() string {
		switch p.(type) {
		case *OpenAIProvider:
			return "*llm.OpenAIProvider"
		case *AnthropicProvider:
			return "*llm.AnthropicProvider"
		default:
			return "unknown"
		}
	}()
}

func TestOpenAIParseStreamLine(t *testing.T) {
	p := &OpenAIProvider{}

	t.Run("done signal", func(t *testing.T) {
		_, isDone, _ := p.ParseStreamLine("data: [DONE]", "")
		if !isDone {
			t.Error("expected isDone=true")
		}
	})

	t.Run("non-data line skipped", func(t *testing.T) {
		_, _, shouldSkip := p.ParseStreamLine("event: something", "")
		if !shouldSkip {
			t.Error("expected shouldSkip=true")
		}
	})

	t.Run("valid data line", func(t *testing.T) {
		data := ResponseData{
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Index        int    `json:"index"`
				FinishReason string `json:"finish_reason"`
			}{
				{Delta: struct {
					Content string `json:"content"`
				}{Content: "hello"}},
			},
		}
		jsonBytes, _ := json.Marshal(data)
		content, isDone, shouldSkip := p.ParseStreamLine("data:"+string(jsonBytes), "")
		if isDone || shouldSkip {
			t.Errorf("isDone=%v, shouldSkip=%v", isDone, shouldSkip)
		}
		if content != "hello" {
			t.Errorf("content = %q, want %q", content, "hello")
		}
	})
}

func TestAnthropicParseStreamLine(t *testing.T) {
	p := &AnthropicProvider{}

	t.Run("message_stop", func(t *testing.T) {
		_, isDone, _ := p.ParseStreamLine("", "message_stop")
		if !isDone {
			t.Error("expected isDone=true")
		}
	})

	t.Run("non-delta event skipped", func(t *testing.T) {
		_, _, shouldSkip := p.ParseStreamLine("", "message_start")
		if !shouldSkip {
			t.Error("expected shouldSkip=true")
		}
	})

	t.Run("content_block_delta", func(t *testing.T) {
		line := `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"world"}}`
		content, isDone, shouldSkip := p.ParseStreamLine(line, "content_block_delta")
		if isDone || shouldSkip {
			t.Errorf("isDone=%v, shouldSkip=%v", isDone, shouldSkip)
		}
		if content != "world" {
			t.Errorf("content = %q, want %q", content, "world")
		}
	})
}

func TestAnthropicBuildRequestBody(t *testing.T) {
	p := &AnthropicProvider{}
	payload := Payload{
		Model: "claude-sonnet-4-5",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "hi"},
		},
		Stream: true,
	}

	body, err := p.BuildRequestBody(payload)
	if err != nil {
		t.Fatal(err)
	}

	req, ok := body.(anthropicRequest)
	if !ok {
		t.Fatal("expected anthropicRequest type")
	}
	if req.System != "You are helpful" {
		t.Errorf("system = %q", req.System)
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
		t.Errorf("messages should have system filtered out, got %d messages", len(req.Messages))
	}
	if req.MaxTokens != 4096 {
		t.Errorf("max_tokens = %d, want 4096", req.MaxTokens)
	}
}
