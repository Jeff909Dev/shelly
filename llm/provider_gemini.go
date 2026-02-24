package llm

import (
	"encoding/json"
	"net/http"
	. "q/types"
	"strings"
)

type GeminiProvider struct{}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig  map[string]interface{} `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiStreamResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) SetHeaders(req *http.Request, auth, orgID string) {
	// Gemini uses query parameter for auth, but also supports header
	req.Header.Set("x-goog-api-key", auth)
}

func (p *GeminiProvider) BuildRequestBody(payload Payload) (interface{}, error) {
	var system *geminiContent
	var contents []geminiContent

	for _, m := range payload.Messages {
		if m.Role == "system" {
			system = &geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			}
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	req := geminiRequest{
		Contents:          contents,
		SystemInstruction: system,
		GenerationConfig: map[string]interface{}{
			"temperature": payload.Temperature,
		},
	}
	return req, nil
}

func (p *GeminiProvider) ParseStreamLine(line, eventType string) (string, bool, bool) {
	// Gemini streams JSON array elements, each prefixed with "data: "
	if !strings.HasPrefix(line, "data:") {
		return "", false, true
	}

	data := strings.TrimPrefix(line, "data: ")
	data = strings.TrimPrefix(data, "data:")

	var resp geminiStreamResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return "", false, true
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", false, true
	}

	return resp.Candidates[0].Content.Parts[0].Text, false, false
}
