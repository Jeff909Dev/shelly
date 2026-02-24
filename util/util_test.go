package util

import "testing"

func TestExtractFirstCodeBlock(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCode   string
		wantIsOnly bool
	}{
		{
			name:       "simple code block",
			input:      "```bash\necho hi\n```",
			wantCode:   "echo hi",
			wantIsOnly: true,
		},
		{
			name:       "code block with surrounding text",
			input:      "Here is the command:\n```bash\nls -la\n```\nThat lists files.",
			wantCode:   "ls -la",
			wantIsOnly: false,
		},
		{
			name:       "no code block",
			input:      "Just some text without code",
			wantCode:   "",
			wantIsOnly: false,
		},
		{
			name:       "empty string",
			input:      "",
			wantCode:   "",
			wantIsOnly: false,
		},
		{
			name:       "short string",
			input:      "hi",
			wantCode:   "",
			wantIsOnly: false,
		},
		{
			name:       "code block no language",
			input:      "```\necho hi\n```",
			wantCode:   "echo hi",
			wantIsOnly: true,
		},
		{
			name:       "multiple code blocks returns first",
			input:      "```bash\nfirst\n```\ntext\n```bash\nsecond\n```",
			wantCode:   "first",
			wantIsOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, isOnly := ExtractFirstCodeBlock(tt.input)
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if isOnly != tt.wantIsOnly {
				t.Errorf("isOnly = %v, want %v", isOnly, tt.wantIsOnly)
			}
		})
	}
}

func TestStartsWithCodeBlock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"starts with backticks", "```bash\necho hi\n```", true},
		{"starts with text", "Hello ```code```", false},
		{"empty", "", false},
		{"just backticks", "```", true},
		{"single backtick", "`", true},
		{"two backticks", "``", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartsWithCodeBlock(tt.input); got != tt.want {
				t.Errorf("StartsWithCodeBlock(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsLikelyBillingError(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"429 Too Many Requests", true},
		{"some other error", false},
		{"error 429 Too Many Requests occurred", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsLikelyBillingError(tt.input); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
