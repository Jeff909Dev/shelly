package llm

import (
	"net/http"
	. "q/types"
	"strings"
)

// Provider abstracts the differences between LLM API providers.
type Provider interface {
	SetHeaders(req *http.Request, auth, orgID string)
	BuildRequestBody(payload Payload) (interface{}, error)
	ParseStreamLine(line, eventType string) (content string, isDone bool, shouldSkip bool)
}

// detectProvider returns the appropriate provider based on the endpoint URL.
func detectProvider(endpoint string) Provider {
	if strings.Contains(endpoint, "anthropic.com") {
		return &AnthropicProvider{}
	}
	return &OpenAIProvider{}
}
