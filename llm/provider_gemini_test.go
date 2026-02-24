package llm

import (
	. "q/types"
	"testing"
)

func TestGeminiDetection(t *testing.T) {
	p := detectProvider("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:streamGenerateContent?alt=sse")
	if _, ok := p.(*GeminiProvider); !ok {
		t.Errorf("expected GeminiProvider, got %T", p)
	}
}

func TestGeminiParseStreamLine(t *testing.T) {
	p := &GeminiProvider{}

	t.Run("valid data line", func(t *testing.T) {
		line := `data: {"candidates":[{"content":{"parts":[{"text":"hello"}]}}]}`
		content, isDone, shouldSkip := p.ParseStreamLine(line, "")
		if isDone || shouldSkip {
			t.Errorf("isDone=%v, shouldSkip=%v", isDone, shouldSkip)
		}
		if content != "hello" {
			t.Errorf("content = %q, want %q", content, "hello")
		}
	})

	t.Run("non-data line skipped", func(t *testing.T) {
		_, _, shouldSkip := p.ParseStreamLine("something else", "")
		if !shouldSkip {
			t.Error("expected shouldSkip=true")
		}
	})

	t.Run("empty candidates skipped", func(t *testing.T) {
		line := `data: {"candidates":[]}`
		_, _, shouldSkip := p.ParseStreamLine(line, "")
		if !shouldSkip {
			t.Error("expected shouldSkip=true for empty candidates")
		}
	})
}

func TestGeminiBuildRequestBody(t *testing.T) {
	p := &GeminiProvider{}
	payload := Payload{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
			{Role: "user", Content: "bye"},
		},
		Stream: true,
	}

	body, err := p.BuildRequestBody(payload)
	if err != nil {
		t.Fatal(err)
	}

	req, ok := body.(geminiRequest)
	if !ok {
		t.Fatal("expected geminiRequest type")
	}

	if req.SystemInstruction == nil {
		t.Fatal("expected system instruction")
	}
	if req.SystemInstruction.Parts[0].Text != "You are helpful" {
		t.Errorf("system = %q", req.SystemInstruction.Parts[0].Text)
	}

	// Should have 3 content messages (system filtered out)
	if len(req.Contents) != 3 {
		t.Fatalf("expected 3 contents, got %d", len(req.Contents))
	}
	if req.Contents[0].Role != "user" {
		t.Errorf("first content role = %q, want user", req.Contents[0].Role)
	}
	if req.Contents[1].Role != "model" {
		t.Errorf("second content role = %q, want model", req.Contents[1].Role)
	}
}
