package llm

import (
	"encoding/json"
	"net/http"
	. "q/types"
	"strings"
)

type AnthropicProvider struct{}

type anthropicRequest struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
}

type anthropicDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicEvent struct {
	Type  string         `json:"type"`
	Delta anthropicDelta `json:"delta"`
}

func (p *AnthropicProvider) SetHeaders(req *http.Request, auth, orgID string) {
	req.Header.Set("x-api-key", auth)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func (p *AnthropicProvider) BuildRequestBody(payload Payload) (interface{}, error) {
	var system string
	var messages []Message
	for _, m := range payload.Messages {
		if m.Role == "system" {
			system = m.Content
		} else {
			messages = append(messages, m)
		}
	}

	maxTokens := payload.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return anthropicRequest{
		Model:     payload.Model,
		System:    system,
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    payload.Stream,
	}, nil
}

func (p *AnthropicProvider) ParseStreamLine(line, eventType string) (string, bool, bool) {
	if eventType == "message_stop" {
		return "", true, false
	}
	if eventType != "content_block_delta" {
		return "", false, true
	}

	// Strip SSE "data: " prefix before parsing JSON
	data := strings.TrimPrefix(line, "data: ")
	data = strings.TrimPrefix(data, "data:")

	var event anthropicEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return "", false, true
	}
	return event.Delta.Text, false, false
}
