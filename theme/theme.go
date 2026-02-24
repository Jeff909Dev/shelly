package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines semantic colors for the UI.
type Theme struct {
	Name    string
	Primary lipgloss.Color
	Error   lipgloss.Color
	Success lipgloss.Color
	Muted   lipgloss.Color
	Accent  lipgloss.Color
}

// Current is the active theme. Set via LoadTheme().
var Current = themes["default"]

// LoadTheme sets the active theme by name. Returns false if not found.
func LoadTheme(name string) bool {
	t, ok := themes[name]
	if !ok {
		return false
	}
	Current = t
	return true
}

// Names returns all available theme names.
func Names() []string {
	return []string{"default", "dracula", "catppuccin-mocha", "nord", "tokyo-night", "gruvbox"}
}
