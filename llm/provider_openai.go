package llm

import (
	"net/http"
	. "q/types"
	"strings"
)

type OpenAIProvider struct{}

func (p *OpenAIProvider) SetHeaders(req *http.Request, auth, orgID string) {
	if strings.Contains(req.URL.String(), "openai.azure.com") {
		req.Header.Set("Api-Key", auth)
	} else {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	if orgID != "" {
		req.Header.Set("OpenAI-Organization", orgID)
	}
}

func (p *OpenAIProvider) BuildRequestBody(payload Payload) (interface{}, error) {
	return payload, nil
}

func (p *OpenAIProvider) ParseStreamLine(line, eventType string) (string, bool, bool) {
	if line == "data: [DONE]" {
		return "", true, false
	}
	if !strings.HasPrefix(line, "data:") {
		return "", false, true
	}
	return strings.TrimPrefix(line, "data:"), false, false
}
