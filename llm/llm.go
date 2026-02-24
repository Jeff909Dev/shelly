package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	. "q/types"
	"strings"
	"time"
)

type LLMClient struct {
	config   ModelConfig
	messages []Message
	provider Provider

	StreamCallback func(string, error)
	Cancel         context.CancelFunc

	httpClient *http.Client
}

func NewLLMClient(config ModelConfig) *LLMClient {
	return &LLMClient{
		config:   config,
		messages: append([]Message(nil), config.Prompt...),
		provider: detectProvider(config.Endpoint),

		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
	}
}

func (c *LLMClient) createRequest(ctx context.Context, payload Payload) (*http.Request, error) {
	body, err := c.provider.BuildRequestBody(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %w", err)
	}
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.provider.SetHeaders(req, c.config.Auth, c.config.OrgID)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *LLMClient) Query(query string) (string, error) {
	messages := c.messages
	messages = append(messages, Message{Role: "user", Content: query})

	payload := Payload{
		Model:       c.config.ModelName,
		Messages:    messages,
		Temperature: 0,
		Stream:      true,
	}

	message, err := c.callStream(payload)
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, message)
	return message.Content, nil
}

func (c *LLMClient) processStream(resp *http.Response) (string, error) {
	counter := 0
	streamReader := bufio.NewReader(resp.Body)
	var buf strings.Builder
	var currentEventType string
	for {
		line, err := streamReader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		// Track SSE event type lines
		if strings.HasPrefix(line, "event:") {
			currentEventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if line == "" {
			continue
		}

		content, isDone, shouldSkip := c.provider.ParseStreamLine(line, currentEventType)
		currentEventType = ""
		if isDone {
			break
		}
		if shouldSkip {
			continue
		}

		if counter < 2 && strings.Count(content, "\n") > 0 {
			continue
		}
		buf.WriteString(content)
		c.StreamCallback(buf.String(), nil)
		counter++
	}
	return buf.String(), nil
}

func (c *LLMClient) callStream(payload Payload) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	c.Cancel = cancel
	req, err := c.createRequest(ctx, payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to create the request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Message{}, fmt.Errorf("API request failed: %s", resp.Status)
	}
	content, err := c.processStream(resp)
	return Message{Role: "assistant", Content: content}, err
}
