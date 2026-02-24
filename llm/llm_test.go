package llm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	. "q/types"
	"strings"
	"testing"
)

func TestProcessStreamOpenAI(t *testing.T) {
	// Simulate an OpenAI SSE stream
	sseData := `data: {"id":"1","choices":[{"delta":{"content":"hello"}}]}

data: {"id":"2","choices":[{"delta":{"content":" world"}}]}

data: {"id":"3","choices":[{"delta":{"content":"!"}}]}

data: [DONE]

`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	config := ModelConfig{
		ModelName: "test-model",
		Endpoint:  server.URL,
		Auth:      "test-key",
	}

	client := NewLLMClient(config)

	var lastPartial string
	client.StreamCallback = func(content string, err error) {
		lastPartial = content
	}

	resp, err := client.Query("test")
	if err != nil {
		t.Fatal(err)
	}

	// The first 2 chunks with newlines are skipped (counter < 2 check)
	// In this case all chunks are single-word so they should all pass
	if !strings.Contains(resp, "hello") {
		t.Errorf("response should contain 'hello', got %q", resp)
	}
	if lastPartial == "" {
		t.Error("stream callback should have been called")
	}
}

func TestProcessStreamAnthropic(t *testing.T) {
	sseData := `event: message_start
data: {"type":"message_start"}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"hi"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" there"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"!"}}

event: message_stop
data: {"type":"message_stop"}

`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	config := ModelConfig{
		ModelName: "claude-test",
		Endpoint:  server.URL + "/anthropic.com/path", // triggers AnthropicProvider detection
		Auth:      "test-key",
	}

	client := NewLLMClient(config)

	var lastPartial string
	client.StreamCallback = func(content string, err error) {
		lastPartial = content
	}

	resp, err := client.Query("test")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(resp, "hi") {
		t.Errorf("response should contain 'hi', got %q", resp)
	}
	if lastPartial == "" {
		t.Error("stream callback should have been called")
	}
}

func TestAPIErrorReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	config := ModelConfig{
		ModelName: "test",
		Endpoint:  server.URL,
		Auth:      "bad-key",
	}

	client := NewLLMClient(config)
	client.StreamCallback = func(string, error) {}

	_, err := client.Query("test")
	if err == nil {
		t.Error("expected error for 401 response")
	}
}
