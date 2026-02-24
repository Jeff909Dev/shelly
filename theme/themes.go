package theme

import "github.com/charmbracelet/lipgloss"

var themes = map[string]Theme{
	"default": {
		Name:    "default",
		Primary: lipgloss.Color("205"),
		Error:   lipgloss.Color("9"),
		Success: lipgloss.Color("2"),
		Muted:   lipgloss.Color("241"),
		Accent:  lipgloss.Color("170"),
	},
	"dracula": {
		Name:    "dracula",
		Primary: lipgloss.Color("#bd93f9"),
		Error:   lipgloss.Color("#ff5555"),
		Success: lipgloss.Color("#50fa7b"),
		Muted:   lipgloss.Color("#6272a4"),
		Accent:  lipgloss.Color("#ff79c6"),
	},
	"catppuccin-mocha": {
		Name:    "catppuccin-mocha",
		Primary: lipgloss.Color("#cba6f7"),
		Error:   lipgloss.Color("#f38ba8"),
		Success: lipgloss.Color("#a6e3a1"),
		Muted:   lipgloss.Color("#6c7086"),
		Accent:  lipgloss.Color("#f5c2e7"),
	},
	"nord": {
		Name:    "nord",
		Primary: lipgloss.Color("#88c0d0"),
		Error:   lipgloss.Color("#bf616a"),
		Success: lipgloss.Color("#a3be8c"),
		Muted:   lipgloss.Color("#4c566a"),
		Accent:  lipgloss.Color("#b48ead"),
	},
	"tokyo-night": {
		Name:    "tokyo-night",
		Primary: lipgloss.Color("#7aa2f7"),
		Error:   lipgloss.Color("#f7768e"),
		Success: lipgloss.Color("#9ece6a"),
		Muted:   lipgloss.Color("#565f89"),
		Accent:  lipgloss.Color("#bb9af7"),
	},
	"gruvbox": {
		Name:    "gruvbox",
		Primary: lipgloss.Color("#fabd2f"),
		Error:   lipgloss.Color("#fb4934"),
		Success: lipgloss.Color("#b8bb26"),
		Muted:   lipgloss.Color("#928374"),
		Accent:  lipgloss.Color("#d3869b"),
	},
}
