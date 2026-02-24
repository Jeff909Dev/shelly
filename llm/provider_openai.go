package llm

import (
	"encoding/json"
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
	raw := strings.TrimPrefix(line, "data:")
	var responseData ResponseData
	if err := json.Unmarshal([]byte(raw), &responseData); err != nil {
		return "", false, true
	}
	if len(responseData.Choices) == 0 {
		return "", false, true
	}
	return responseData.Choices[0].Delta.Content, false, false
}
