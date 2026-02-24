package llm

import (
	"fmt"
	"net/http"
	"q/plugin"
	. "q/types"
	"strings"
)

// Provider abstracts the differences between LLM API providers.
type Provider interface {
	SetHeaders(req *http.Request, auth, orgID string)
	BuildRequestBody(payload Payload) (interface{}, error)
	ParseStreamLine(line, eventType string) (content string, isDone bool, shouldSkip bool)
}

// DetectProvider returns the appropriate provider based on the endpoint URL or plugin name.
func DetectProvider(config ModelConfig) (Provider, error) {
	if config.Plugin != "" {
		manifests, err := plugin.Discover()
		if err != nil {
			return nil, fmt.Errorf("failed to discover plugins: %w", err)
		}
		for _, m := range manifests {
			if m.Name == config.Plugin && m.Type == "provider" {
				return plugin.NewPluginProvider(m.Executable)
			}
		}
		return nil, fmt.Errorf("plugin %q not found", config.Plugin)
	}
	return detectProvider(config.Endpoint), nil
}

// detectProvider returns the appropriate provider based on the endpoint URL.
func detectProvider(endpoint string) Provider {
	if strings.Contains(endpoint, "anthropic.com") {
		return &AnthropicProvider{}
	}
	if strings.Contains(endpoint, "generativelanguage.googleapis.com") {
		return &GeminiProvider{}
	}
	return &OpenAIProvider{}
}
