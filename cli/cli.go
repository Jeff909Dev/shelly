package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"q/config"
	"q/history"
	"q/llm"
	"q/theme"
	. "q/types"
	"q/util"

	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type State int

const (
	Loading State = iota
	ReceivingInput
	ReceivingResponse
)

type model struct {
	client           *llm.LLMClient
	markdownRenderer *glamour.TermRenderer
	p                *tea.Program

	textInput textinput.Model
	spinner   spinner.Model

	state                 State
	query                 string
	latestCommandResponse string
	latestCommandIsCode   bool

	formattedPartialResponse string

	maxWidth int

	runWithArgs    bool
	err            error
	historyStore   *history.HistoryStore
	historyEnabled bool
	modelName      string
}

type responseMsg struct {
	response string
	err      error
}
type partialResponseMsg struct {
	content string
	err     error
}
type setPMsg struct{ p *tea.Program }

// === Commands === //

func makeQuery(client *llm.LLMClient, query string) tea.Cmd {
	return func() tea.Msg {
		response, err := client.Query(query)
		return responseMsg{response: response, err: err}
	}
}

// === Msg Handlers === //

func (m model) handleKeyEnter() (tea.Model, tea.Cmd) {
	if m.state != ReceivingInput {
		return m, nil
	}
	v := m.textInput.Value()

	// No input, copy and quit.
	if v == "" {
		if m.latestCommandResponse == "" {
			return m, tea.Quit
		}
		err := clipboard.WriteAll(m.latestCommandResponse)
		if err != nil {
			fmt.Println("Failed to copy text to clipboard:", err)
			return m, tea.Quit
		}
		placeholderStyle := lipgloss.NewStyle().Faint(true)
		message := "Copied to clipboard."
		if !m.latestCommandIsCode {
			message = "Copied only code to clipboard."
		}
		message = placeholderStyle.Render(message)
		return m, tea.Sequence(tea.Printf("%s", message), tea.Quit)
	}
	// Input, run query.
	m.textInput.SetValue("")
	m.query = v
	m.state = Loading
	placeholderStyle := lipgloss.NewStyle().Faint(true).Width(m.maxWidth)
	message := placeholderStyle.Render(fmt.Sprintf("> %s", v))
	return m, tea.Sequence(tea.Printf("%s", message), tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query)))
}

func (m model) formatResponse(response string, isCode bool) (string, error) {

	// format nicely
	formatted, err := m.markdownRenderer.Render(response)
	if err != nil {
		return response, nil
	}

	// trim preceding and trailing newlines
	formatted = strings.TrimPrefix(formatted, "\n")
	formatted = strings.TrimSuffix(formatted, "\n")

	// Add newline for non-code blocks (hacky)
	if !isCode {
		formatted = "\n" + formatted
	}
	return formatted, nil
}

// TODO: parse the model endpoint to infer whether it's openai, other, or local.
// for local, suggest it may not be running, and how to run it
func (m model) getConnectionError(err error) string {
	styleRed := lipgloss.NewStyle().Foreground(theme.Current.Error)
	styleGreen := lipgloss.NewStyle().Foreground(theme.Current.Success)
	styleDim := lipgloss.NewStyle().Faint(true).Width(m.maxWidth).PaddingLeft(2)
	message := fmt.Sprintf("\n  %v\n\n%v\n",
		styleRed.Render("Error: Failed to connect to LLM API."),
		styleDim.Render(err.Error()))
	if util.IsLikelyBillingError(err.Error()) {
		message = fmt.Sprintf("%v\n  %v %v\n\n  %v%v\n\n",
			message,
			styleGreen.Render("Hint:"),
			"You may need to set up billing. You can do so here:",
			styleGreen.Render("->"),
			styleDim.Render("https://platform.openai.com/account/billing"),
		)
	}
	return message
}

func (m model) handleResponseMsg(msg responseMsg) (tea.Model, tea.Cmd) {
	m.formattedPartialResponse = ""

	// error handling
	if msg.err != nil {
		m.state = ReceivingInput
		message := m.getConnectionError(msg.err)
		return m, tea.Sequence(tea.Printf("%s", message), textinput.Blink)
	}

	// parse out the code block
	content, isOnlyCode := util.ExtractFirstCodeBlock(msg.response)
	if content != "" {
		m.latestCommandResponse = content
	}

	formatted, _ := m.formatResponse(msg.response, util.StartsWithCodeBlock(msg.response))

	m.textInput.Placeholder = "Follow up, ENTER to copy & quit, CTRL+C to quit"
	if !isOnlyCode {
		m.textInput.Placeholder = "Follow up, ENTER to copy (code only), CTRL+C to quit"
	}
	if m.latestCommandResponse == "" {
		m.textInput.Placeholder = "Follow up, ENTER or CTRL+C to quit"
	}

	m.state = ReceivingInput
	m.latestCommandIsCode = isOnlyCode

	// Save to history
	if m.historyEnabled && m.historyStore != nil {
		id := generateID()
		_ = m.historyStore.Save(history.Conversation{
			ID:        id,
			Timestamp: time.Now(),
			Model:     m.modelName,
			Messages: []Message{
				{Role: "user", Content: m.query},
				{Role: "assistant", Content: msg.response},
			},
		})
	}

	message := formatted
	return m, tea.Sequence(tea.Printf("%s", message), textinput.Blink)
}

func (m model) handlePartialResponseMsg(msg partialResponseMsg) (tea.Model, tea.Cmd) {
	m.state = ReceivingResponse
	isCode := util.StartsWithCodeBlock(msg.content)
	formatted, _ := m.formatResponse(msg.content, isCode)
	m.formattedPartialResponse = formatted
	return m, nil
}

// === Init, Update, View === //

func (m model) Init() tea.Cmd {
	if m.runWithArgs {
		return tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query))
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleKeyEnter()
		}

	case responseMsg:
		return m.handleResponseMsg(msg)

	case partialResponseMsg:
		return m.handlePartialResponseMsg(msg)

	case setPMsg:
		m.p = msg.p
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}
	// Update spinner or cursor.
	switch m.state {
	case Loading:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case ReceivingInput:
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case Loading:
		return m.spinner.View()
	case ReceivingInput:
		return m.textInput.View()
	case ReceivingResponse:
		return m.formattedPartialResponse + "\n"
	}
	return ""
}

// === Initial Model Setup === //

func initialModel(prompt string, client *llm.LLMClient, histStore *history.HistoryStore, histEnabled bool, modelName string) model {
	maxWidth := util.GetTermSafeMaxWidth()
	ti := textinput.New()
	ti.Placeholder = "Describe a shell command, or ask a question."
	ti.Focus()
	ti.Width = maxWidth

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Current.Primary)

	runWithArgs := prompt != ""

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(int(maxWidth)),
	)
	model := model{
		client:                client,
		markdownRenderer:      r,
		textInput:             ti,
		spinner:               s,
		state:                 ReceivingInput,
		query:                 "",
		latestCommandResponse: "",
		latestCommandIsCode:   false,
		maxWidth:              maxWidth,
		runWithArgs:           false,
		err:                   nil,
		historyStore:          histStore,
		historyEnabled:        histEnabled,
		modelName:             modelName,
	}

	if runWithArgs {
		model.runWithArgs = true
		model.state = Loading
		model.query = prompt
	}
	return model
}

// === Main === //

func streamHandler(p *tea.Program) func(content string, err error) {
	return func(content string, err error) {
		p.Send(partialResponseMsg{content, err})
	}
}

func getModelConfig(appConfig config.AppConfig) (ModelConfig, error) {
	if len(appConfig.Models) == 0 {
		return ModelConfig{}, fmt.Errorf("no models available")
	}
	for _, model := range appConfig.Models {
		if model.ModelName == appConfig.Preferences.DefaultModel {
			return model, nil
		}
	}
	// If the preferred model is not found, return the first model
	return appConfig.Models[0], nil
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func readStdin() string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return ""
	}
	// Read piped input, limit to 100KB
	const maxSize = 100 * 1024
	buf := make([]byte, maxSize)
	n, _ := os.Stdin.Read(buf)
	if n == 0 {
		return ""
	}
	content := string(buf[:n])
	// Reject likely binary content
	for _, b := range content[:min(512, len(content))] {
		if b == 0 {
			return ""
		}
	}
	return content
}

func runQProgram(prompt string) {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		config.PrintConfigErrorMessage(err)
		os.Exit(1)
	}

	if appConfig.Preferences.Theme != "" {
		theme.LoadTheme(appConfig.Preferences.Theme)
	}

	modelConfig, err := getModelConfig(appConfig)
	if err != nil {
		config.PrintConfigErrorMessage(err)
		os.Exit(1)
	}
	auth := os.Getenv(modelConfig.Auth)
	if auth == "" {
		envVar, key := config.RunSetupWizard(appConfig)
		if envVar == "" || key == "" {
			return
		}
		os.Setenv(envVar, key)
		// Reload config since wizard may have changed default model
		appConfig, _ = config.LoadAppConfig()
		modelConfig, err = getModelConfig(appConfig)
		if err != nil {
			config.PrintConfigErrorMessage(err)
			os.Exit(1)
		}
		auth = key
	}
	// everything checks out, save the config
	// TODO: maybe add a validating function
	config.SaveAppConfig(appConfig)

	orgID := os.Getenv(modelConfig.OrgID)
	modelConfig.Auth = auth
	modelConfig.OrgID = orgID

	// Set up history
	var histStore *history.HistoryStore
	histEnabled := appConfig.Preferences.HistoryEnabled
	if histEnabled {
		histStore, _ = history.NewStore()
		if histStore != nil && appConfig.Preferences.HistoryMaxDays > 0 {
			_ = histStore.Prune(appConfig.Preferences.HistoryMaxDays)
		}
	}

	c := llm.NewLLMClient(modelConfig)
	p := tea.NewProgram(initialModel(prompt, c, histStore, histEnabled, modelConfig.ModelName))
	c.StreamCallback = streamHandler(p)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View conversation history",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := history.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		convs, err := store.List(20)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(convs) == 0 {
			fmt.Println("No history yet.")
			return
		}
		for _, c := range convs {
			query := c.Messages[0].Content
			if len(query) > 60 {
				query = query[:60] + "..."
			}
			fmt.Printf("  %s  %s  %s  %s\n",
				lipgloss.NewStyle().Foreground(theme.Current.Muted).Render(c.ID),
				lipgloss.NewStyle().Foreground(theme.Current.Muted).Render(c.Timestamp.Format("2006-01-02 15:04")),
				lipgloss.NewStyle().Foreground(theme.Current.Accent).Render(c.Model),
				query,
			)
		}
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search conversation history",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		store, err := history.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		results, err := store.Search(strings.Join(args, " "))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(results) == 0 {
			fmt.Println("No matches found.")
			return
		}
		for _, c := range results {
			query := c.Messages[0].Content
			if len(query) > 60 {
				query = query[:60] + "..."
			}
			fmt.Printf("  %s  %s  %s\n",
				lipgloss.NewStyle().Foreground(theme.Current.Muted).Render(c.ID),
				lipgloss.NewStyle().Foreground(theme.Current.Accent).Render(c.Model),
				query,
			)
		}
	},
}

var historyShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a conversation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		store, err := history.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conv, err := store.Show(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ID: %s  Model: %s  Time: %s\n\n",
			conv.ID, conv.Model, conv.Timestamp.Format("2006-01-02 15:04:05"))
		for _, msg := range conv.Messages {
			role := lipgloss.NewStyle().Bold(true).Render(msg.Role + ":")
			fmt.Printf("%s %s\n\n", role, msg.Content)
		}
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all history",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := history.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := store.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("History cleared.")
	},
}

func runSuggest(prompt string) {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		os.Exit(1)
	}
	modelConfig, err := getModelConfig(appConfig)
	if err != nil {
		os.Exit(1)
	}
	auth := os.Getenv(modelConfig.Auth)
	if auth == "" {
		os.Exit(1)
	}
	orgID := os.Getenv(modelConfig.OrgID)
	modelConfig.Auth = auth
	modelConfig.OrgID = orgID

	// Override prompt to completion mode
	modelConfig.Prompt = []Message{
		{Role: "system", Content: "Complete this partial shell command. Output ONLY the completed command, nothing else. No markdown, no code blocks, no explanation."},
	}

	c := llm.NewLLMClient(modelConfig)
	// Use a no-op stream callback
	c.StreamCallback = func(s string, e error) {}
	result, err := c.Query(prompt)
	if err != nil {
		os.Exit(1)
	}
	// Strip any accidental markdown
	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "```bash\n")
	result = strings.TrimPrefix(result, "```\n")
	result = strings.TrimSuffix(result, "\n```")
	result = strings.TrimSuffix(result, "```")
	fmt.Print(result)
}

func init() {
	historyCmd.AddCommand(historySearchCmd, historyShowCmd, historyClearCmd)
	RootCmd.AddCommand(historyCmd)
	RootCmd.Flags().Bool("suggest", false, "Output a command completion suggestion (for shell integration)")
}

var RootCmd = &cobra.Command{
	Use:   "q [request]",
	Short: "A command line interface for natural language queries",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// join args into a single string separated by spaces
		prompt := strings.Join((args), " ")
		if len(args) > 0 && args[0] == "config" {
			config.RunConfigProgram(args)
			return
		}

		// Check for piped stdin
		stdinContent := readStdin()
		if stdinContent != "" {
			prompt = fmt.Sprintf("Input:\n```\n%s\n```\n\n%s", stdinContent, prompt)
		}

		// Inject shell context
		shellCtx := util.GetShellContext()
		if shellCtx != "" {
			prompt = shellCtx + "\n" + prompt
		}

		suggest, _ := cmd.Flags().GetBool("suggest")
		if suggest {
			runSuggest(prompt)
			return
		}

		runQProgram(prompt)
	},
}
