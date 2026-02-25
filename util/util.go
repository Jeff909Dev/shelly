package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/mattn/go-tty"
)

const (
	TermMaxWidth        = 100
	TermSafeZonePadding = 10
)

func StartsWithCodeBlock(s string) bool {
	n := len(s)
	if n == 0 {
		return false
	}
	if n <= 3 {
		for i := 0; i < n; i++ {
			if s[i] != '`' {
				return false
			}
		}
		return true
	}
	return s[0] == '`' && s[1] == '`' && s[2] == '`'
}

func ExtractFirstCodeBlock(s string) (content string, isOnlyCode bool) {
	isOnlyCode = true
	if len(s) <= 3 {
		return "", false
	}
	start := strings.Index(s, "```")
	if start == -1 {
		return "", false
	}
	if start != 0 {
		isOnlyCode = false
	}
	fromStart := s[start:]
	content = strings.TrimPrefix(fromStart, "```")
	// Find newline after the first ```
	newlinePos := strings.Index(content, "\n")
	if newlinePos != -1 {
		// Check if there's a word immediately after the first ```
		if content[0:newlinePos] == strings.TrimSpace(content[0:newlinePos]) {
			// If so, remove that part from the content
			content = content[newlinePos+1:]
		}
	}
	// Strip final ``` if present
	end := strings.Index(content, "```")
	if end < len(content)-3 {
		isOnlyCode = false
	}
	if end != -1 {
		content = content[:end]
	}
	if len(content) == 0 {
		return "", false
	}
	// Strip the final newline, if present
	if content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	return
}

func GetTermSafeMaxWidth() int {
	maxWidth := TermMaxWidth
	termWidth, err := getTermWidth()
	if err != nil {
		return maxWidth
	}
	if termWidth < maxWidth {
		maxWidth = termWidth - TermSafeZonePadding
	}
	if maxWidth < 20 {
		maxWidth = 20
	}
	return maxWidth
}

func getTermWidth() (width int, err error) {
	t, err := tty.Open()
	if err != nil {
		return 0, err
	}
	defer t.Close()
	width, _, err = t.Size()
	return width, err
}

func IsLikelyBillingError(s string) bool {
	return strings.Contains(s, "429 Too Many Requests")
}

func GetShellContext() string {
	var parts []string

	// Current directory
	if cwd, err := os.Getwd(); err == nil {
		parts = append(parts, "cwd: "+cwd)
	}

	// Git branch
	if branch, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		b := strings.TrimSpace(string(branch))
		if b != "" {
			parts = append(parts, "branch: "+b)
		}
	}

	// File listing (limit 30)
	if entries, err := os.ReadDir("."); err == nil {
		var names []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			name := e.Name()
			if e.IsDir() {
				name += "/"
			}
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) > 30 {
			names = names[:30]
		}
		if len(names) > 0 {
			parts = append(parts, "files: "+strings.Join(names, " "))
		}
	}

	// Git status
	if status, err := exec.Command("git", "status", "--short").Output(); err == nil {
		s := strings.TrimSpace(string(status))
		if s != "" {
			// Compact: join lines with comma
			lines := strings.Split(s, "\n")
			if len(lines) > 10 {
				lines = lines[:10]
			}
			parts = append(parts, "git: "+strings.Join(lines, ", "))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("[Context] %s", strings.Join(parts, " | "))
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // For Linux or anything else
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
