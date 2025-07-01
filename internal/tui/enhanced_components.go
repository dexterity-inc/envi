package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dexterity-inc/envi/internal/config"
)

// Package tui provides terminal user interface components for the envi CLI.

// ProgressModel shows progress for operations
type ProgressModel struct {
	progress progress.Model
	spinner  spinner.Model
	message  string
	percent  float64
	done     bool
}

// NewProgressModel creates a new progress model
func NewProgressModel() ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(enhancedPrimary)

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		progress: p,
		spinner:  s,
		message:  "Initializing...",
		percent:  0,
		done:     false,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case spinner.TickMsg:
		spinnerModel, cmd := m.spinner.Update(msg)
		m.spinner = spinnerModel
		return m, cmd
	}
	return m, nil
}

func (m ProgressModel) View() string {
	if m.done {
		return ContainerStyle.Render(
			TitleStyle.Render("Complete") + "\n" +
				SuccessStyle.Render("✅ "+m.message) + "\n" +
				"Press any key to continue...",
		)
	}

	return ContainerStyle.Render(
		TitleStyle.Render("Processing") + "\n\n" +
			m.spinner.View() + " " + m.message + "\n\n" +
			m.progress.View() + "\n\n" +
			"Press q to quit",
	)
}

// ShowProgress displays a progress indicator
func ShowProgress(message string, percent float64) {
	model := NewProgressModel()
	model.message = message
	model.percent = percent

	program := tea.NewProgram(&model)
	if _, err := program.Run(); err != nil {
		fmt.Printf("%s: %.1f%%\n", message, percent*100)
	}
}

// ShowProgressWithCompletion shows progress and completion
func ShowProgressWithCompletion(message string, percent float64, done bool) {
	model := NewProgressModel()
	model.message = message
	model.percent = percent
	model.done = done

	program := tea.NewProgram(&model)
	if _, err := program.Run(); err != nil {
		if done {
			fmt.Printf("✅ %s\n", message)
		} else {
			fmt.Printf("%s: %.1f%%\n", message, percent*100)
		}
	}
}

// ErrorDisplay shows styled error messages with suggestions
type ErrorDisplay struct {
	error       error
	suggestions []string
	nextSteps   []string
	quitting    bool
}

// NewErrorDisplay creates a new error display
func NewErrorDisplay(err error, suggestions []string, nextSteps []string) *ErrorDisplay {
	return &ErrorDisplay{
		error:       err,
		suggestions: suggestions,
		nextSteps:   nextSteps,
	}
}

func (e *ErrorDisplay) Init() tea.Cmd {
	return nil
}

func (e *ErrorDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			e.quitting = true
			return e, tea.Quit
		}
	}
	return e, nil
}

func (e *ErrorDisplay) View() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render("Error") + "\n\n")
	sb.WriteString(ErrorStyle.Render("❌ "+e.error.Error()) + "\n\n")

	if len(e.suggestions) > 0 {
		sb.WriteString(InfoStyle.Render("Suggestions:") + "\n")
		for _, suggestion := range e.suggestions {
			sb.WriteString("• " + suggestion + "\n")
		}
		sb.WriteString("\n")
	}

	if len(e.nextSteps) > 0 {
		sb.WriteString(InfoStyle.Render("Next Steps:") + "\n")
		for _, step := range e.nextSteps {
			sb.WriteString("• " + step + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(HelpStyle.Render("Press any key to continue..."))

	return ContainerStyle.Render(sb.String())
}

// ShowError displays an error with context
func ShowError(err error, context string) {
	suggestions := getErrorSuggestions(err, context)
	nextSteps := getNextSteps(err, context)
	model := NewErrorDisplay(err, suggestions, nextSteps)
	program := tea.NewProgram(model)
	program.Run()
}

func getErrorSuggestions(err error, context string) []string {
	// Add context-specific suggestions here
	return []string{
		"Check your internet connection",
		"Verify your GitHub credentials",
		"Ensure you have the necessary permissions",
	}
}

func getNextSteps(err error, context string) []string {
	// Add context-specific next steps here
	return []string{
		"Run 'envi config' to check your configuration",
		"Try running the command again",
		"Check the documentation for more information",
	}
}

// SuccessDisplay shows success messages
type SuccessDisplay struct {
	content  string
	details  []string
	quitting bool
}

func (s *SuccessDisplay) Init() tea.Cmd {
	return nil
}

func (s *SuccessDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			s.quitting = true
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *SuccessDisplay) View() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render("Success") + "\n\n")
	sb.WriteString(SuccessStyle.Render("✅ "+s.content) + "\n\n")

	if len(s.details) > 0 {
		for _, detail := range s.details {
			sb.WriteString("• " + detail + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(HelpStyle.Render("Press any key to continue..."))

	return ContainerStyle.Render(sb.String())
}

// ShowSuccess displays a success message
func ShowSuccess(message string, details []string) {
	model := &SuccessDisplay{
		content: message,
		details: details,
	}
	program := tea.NewProgram(model)
	program.Run()
}

// InfoDisplay shows informational messages
type InfoDisplay struct {
	title    string
	content  string
	quitting bool
}

func (i *InfoDisplay) Init() tea.Cmd {
	return nil
}

func (i *InfoDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			i.quitting = true
			return i, tea.Quit
		}
	}
	return i, nil
}

func (i *InfoDisplay) View() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(i.title) + "\n\n")
	sb.WriteString(i.content + "\n\n")
	sb.WriteString(HelpStyle.Render("Press any key to continue..."))

	return ContainerStyle.Render(sb.String())
}

// ShowInfo displays an informational message
func ShowInfo(title, message string) {
	model := &InfoDisplay{
		title:   title,
		content: message,
	}
	program := tea.NewProgram(model)
	program.Run()
}

// EnhancedConfirm shows a confirmation dialog with multiple options
type EnhancedConfirm struct {
	title    string
	question string
	options  []string
	selected int
	quitting bool
}

func (c *EnhancedConfirm) Init() tea.Cmd {
	return nil
}

func (c *EnhancedConfirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.selected > 0 {
				c.selected--
			}
		case "down", "j":
			if c.selected < len(c.options)-1 {
				c.selected++
			}
		case "enter":
			c.quitting = true
			return c, tea.Quit
		case "q", "esc", "ctrl+c":
			c.selected = -1
			c.quitting = true
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c *EnhancedConfirm) View() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(c.title) + "\n\n")
	sb.WriteString(c.question + "\n\n")

	for i, option := range c.options {
		if i == c.selected {
			sb.WriteString(SelectedStyle.Render("▶ "+option) + "\n")
		} else {
			sb.WriteString("  " + option + "\n")
		}
	}

	sb.WriteString("\n" + HelpStyle.Render("Use ↑/↓ to navigate, Enter to select, Esc to cancel"))

	return ContainerStyle.Render(sb.String())
}

// ConfirmEnhanced shows an enhanced confirmation dialog
func ConfirmEnhanced(title, question string, options []string) (int, error) {
	model := &EnhancedConfirm{
		title:    title,
		question: question,
		options:  options,
		selected: 0,
	}
	program := tea.NewProgram(model)
	_, err := program.Run()
	return model.selected, err
}

// GistManager key bindings
type gistKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Enter    key.Binding
	Search   key.Binding
	Details  key.Binding
	CopyID   key.Binding
	CopyURL  key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func (k gistKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Search, k.Help, k.Quit}
}

func (k gistKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Enter, k.Search, k.Details, k.CopyID},
		{k.CopyURL, k.Filter, k.Help, k.Quit},
	}
}

// AllKeyBindings returns a comprehensive list of all working key bindings
func (k gistKeyMap) AllKeyBindings() string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("Gist Manager - Key Bindings") + "\n\n")

	// Navigation section
	sb.WriteString(InfoStyle.Render("Navigation") + "\n")
	sb.WriteString("  ↑ / k               : Move up\n")
	sb.WriteString("  ↓ / j               : Move down\n")
	sb.WriteString("  PgUp                : Page up\n")
	sb.WriteString("  PgDn                : Page down\n\n")

	// Actions section
	sb.WriteString(InfoStyle.Render("Actions") + "\n")
	sb.WriteString("  Enter               : Open selected gist in browser\n")
	sb.WriteString("  /                   : Search gists\n")
	sb.WriteString("  d                   : Toggle details view\n")
	sb.WriteString("  y                   : Copy gist ID to clipboard\n")
	sb.WriteString("  Y                   : Copy gist URL to clipboard\n\n")

	// Filters section
	sb.WriteString(InfoStyle.Render("Filters") + "\n")
	sb.WriteString("  f                   : Show all gists\n")
	sb.WriteString("  e                   : Show encrypted gists only\n")
	sb.WriteString("  p                   : Show public gists only\n")
	sb.WriteString("  r                   : Show recent gists only\n\n")

	// Other section
	sb.WriteString(InfoStyle.Render("Other") + "\n")
	sb.WriteString("  h / ?               : Show/hide this help\n")
	sb.WriteString("  q / Esc / Ctrl+C    : Quit\n\n")

	// Search section
	sb.WriteString(InfoStyle.Render("Search") + "\n")
	sb.WriteString("  • Multi-word queries supported\n")
	sb.WriteString("  • Case-insensitive search\n")
	sb.WriteString("  • Searches across name, description, project, environment, ID, and tags\n")
	sb.WriteString("  • Press Enter to confirm search\n")
	sb.WriteString("  • Press Esc to clear search\n")

	return sb.String()
}

// GistManager manages the interactive gist list for gists.
type GistManager struct {
	gists       []*config.GistInfo
	filtered    []*config.GistInfo
	selected    int
	quitting    bool
	width       int
	height      int
	ready       bool
	showDetails bool
	searchTerm  string
	filter      string // "all", "encrypted", "public", "recent"

	// Search input
	searchInput   textinput.Model
	searchFocused bool

	// Pagination
	currentPage  int
	itemsPerPage int

	// UI state
	copyMessage     string
	copyMessageTime time.Time

	// Help
	help         help.Model
	keys         gistKeyMap
	showFullHelp bool
}

const defaultItemsPerPage = 15

// NewGistManager creates a new GistManager with universal key bindings.
func NewGistManager(gists []*config.GistInfo) *GistManager {
	// Initialize search input with better configuration
	searchInput := textinput.New()
	searchInput.Placeholder = "Search gists by name, project, or environment..."
	searchInput.Prompt = "🔍 "
	searchInput.Width = 40
	searchInput.CharLimit = 100

	// Initialize key bindings with only working combinations
	keys := gistKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("PgDn", "page down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "open"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Details: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "details"),
		),
		CopyID: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy id"),
		),
		CopyURL: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "copy url"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f", "e", "p", "r"),
			key.WithHelp("f/e/p/r", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("h", "?"),
			key.WithHelp("h/?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/Esc/Ctrl+C", "quit"),
		),
	}

	gm := &GistManager{
		gists:        gists,
		filtered:     make([]*config.GistInfo, 0, len(gists)),
		selected:     0,
		filter:       "all",
		searchInput:  searchInput,
		itemsPerPage: defaultItemsPerPage,
		help:         help.New(),
		keys:         keys,
		showFullHelp: false,
	}

	// Initialize filtered list properly
	gm.applyFilters()

	return gm
}

// Init initializes the GistManager model.
func (g *GistManager) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the GistManager state.
func (g *GistManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode
		if g.searchFocused {
			return g.handleSearchInput(msg)
		}

		// Handle main navigation and actions
		return g.handleMainInput(msg)

	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		g.ready = true
		g.adjustItemsPerPage()
	}

	return g, cmd
}

// handleSearchInput handles input when search is focused
func (g *GistManager) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc", "ctrl+c":
		g.searchFocused = false
		g.searchInput.Blur()
		g.searchTerm = ""
		g.applyFilters()
		return g, nil
	case "enter":
		g.searchFocused = false
		g.searchInput.Blur()
		g.searchTerm = g.searchInput.Value()
		g.applyFilters()
		return g, nil
	default:
		g.searchInput, cmd = g.searchInput.Update(msg)
		// Only apply filters if the search term actually changed
		newSearchTerm := g.searchInput.Value()
		if newSearchTerm != g.searchTerm {
			g.searchTerm = newSearchTerm
			g.applyFilters()
		}
		return g, cmd
	}
}

// handleMainInput handles input when not in search mode
func (g *GistManager) handleMainInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// Quit
	case "q", "esc", "ctrl+c":
		g.quitting = true
		return g, tea.Quit

	// Navigation
	case "up", "k":
		g.moveSelection(-1)
	case "down", "j":
		g.moveSelection(1)
	case "pgup":
		g.pageUp()
	case "pgdown":
		g.pageDown()

	// Actions
	case "enter":
		g.openSelectedGist()
	case "/":
		g.searchFocused = true
		g.searchInput.Focus()
		return g, textinput.Blink
	case "d":
		g.showDetails = !g.showDetails
	case "y":
		g.copySelectedGistID()
	case "Y":
		g.copySelectedGistURL()

		// Filters
	case "f":
		g.filter = "all"
		g.applyFilters()
	case "e":
		g.filter = "encrypted"
		g.applyFilters()
	case "p":
		g.filter = "public"
		g.applyFilters()
	case "r":
		g.filter = "recent"
		g.applyFilters()

	// Help
	case "h", "?":
		g.showFullHelp = !g.showFullHelp
	}

	return g, nil
}

func (g *GistManager) applyFilters() {
	// Pre-allocate slice with estimated capacity to reduce memory allocations
	estimatedSize := len(g.gists) / 2 // Estimate that roughly half will match filters
	if estimatedSize < 10 {
		estimatedSize = len(g.gists) // For small lists, allocate full capacity
	}
	g.filtered = make([]*config.GistInfo, 0, estimatedSize)

	for _, gist := range g.gists {
		// Apply filter
		if !g.matchesFilter(gist) {
			continue
		}

		// Apply search
		if g.searchTerm != "" && !g.matchesSearch(gist) {
			continue
		}

		// Since gist IDs should be unique, we don't need duplicate checking
		g.filtered = append(g.filtered, gist)
	}

	// Reset pagination and selection
	g.currentPage = 0
	if len(g.filtered) > 0 {
		if g.selected >= len(g.filtered) {
			g.selected = len(g.filtered) - 1
		}
	} else {
		g.selected = 0
	}
}

func (g *GistManager) matchesFilter(gist *config.GistInfo) bool {
	switch g.filter {
	case "encrypted":
		return gist.IsEncrypted
	case "public":
		return gist.IsPublic
	case "recent":
		if gist.LastUsed == "" {
			return false
		}
		weekAgo := time.Now().AddDate(0, 0, -7)
		lastUsed, err := time.Parse("2006-01-02 15:04:05", gist.LastUsed)
		return err == nil && lastUsed.After(weekAgo)
	default:
		return true
	}
}

func (g *GistManager) matchesSearch(gist *config.GistInfo) bool {
	if g.searchTerm == "" {
		return true
	}

	// Convert search term to lowercase once
	lowerSearchTerm := strings.ToLower(strings.TrimSpace(g.searchTerm))

	// Split search terms by spaces for multi-word search
	terms := strings.Fields(lowerSearchTerm)

	// Pre-build searchable text once to avoid repeated allocations
	searchableText := strings.ToLower(gist.Name + " " +
		gist.Description + " " +
		gist.ProjectName + " " +
		gist.Environment + " " +
		gist.ID + " " +
		strings.Join(gist.Tags, " "))

	// All terms must match the combined searchable text
	for _, term := range terms {
		if !strings.Contains(searchableText, term) {
			return false
		}
	}

	return true
}

func (g *GistManager) moveSelection(delta int) {
	if len(g.filtered) == 0 {
		return
	}
	pageStart := g.currentPage * g.itemsPerPage
	pageEnd := pageStart + g.itemsPerPage
	if pageEnd > len(g.filtered) {
		pageEnd = len(g.filtered)
	}
	pageLen := pageEnd - pageStart
	localIdx := g.selected - pageStart
	newLocal := localIdx + delta
	if newLocal < 0 {
		newLocal = pageLen - 1
	} else if newLocal >= pageLen {
		newLocal = 0
	}
	g.selected = pageStart + newLocal
}

func (g *GistManager) pageUp() {
	if g.currentPage > 0 {
		g.currentPage--
		g.selected = g.currentPage * g.itemsPerPage
	}
}

func (g *GistManager) pageDown() {
	maxPage := g.totalPages() - 1
	if g.currentPage < maxPage {
		g.currentPage++
		g.selected = g.currentPage * g.itemsPerPage
	}
}

func (g *GistManager) ensureSelectionInCurrentPage() {
	start := g.currentPage * g.itemsPerPage
	end := start + g.itemsPerPage

	if g.selected < start {
		g.currentPage = g.selected / g.itemsPerPage
	} else if g.selected >= end {
		g.currentPage = g.selected / g.itemsPerPage
	}
}

func (g *GistManager) renderLeftPane(width, height int) string {
	var sb strings.Builder
	pageItems := g.getCurrentPageItems()
	sb.WriteString(g.renderGistList(pageItems, height-3))
	if g.totalPages() > 1 {
		sb.WriteString(fmt.Sprintf("\nPage %d/%d (%d total gists)", g.currentPage+1, g.totalPages(), len(g.filtered)))
	}
	return lipgloss.NewStyle().Width(width).Render(sb.String())
}

func (g *GistManager) renderRightPane(width, height int) string {
	var sb strings.Builder
	if len(g.filtered) > 0 && g.selected < len(g.filtered) {
		selectedGist := g.filtered[g.selected]
		// Remove the duplicate title since renderGistDetails has its own sections
		sb.WriteString(g.renderGistDetails(selectedGist))
	} else {
		sb.WriteString(InfoStyle.Render("No gist selected."))
	}
	return lipgloss.NewStyle().Width(width).Render(sb.String())
}

func (g *GistManager) adjustItemsPerPage() {
	// Reserve space for header, search, filters, help, and margins
	availableHeight := g.height - 10
	g.itemsPerPage = max(5, min(availableHeight, defaultItemsPerPage))
}

func (g *GistManager) totalPages() int {
	if len(g.filtered) == 0 {
		return 1
	}
	return (len(g.filtered)-1)/g.itemsPerPage + 1
}

func (g *GistManager) getCurrentPageItems() []*config.GistInfo {
	start := g.currentPage * g.itemsPerPage
	end := min(start+g.itemsPerPage, len(g.filtered))

	if start >= len(g.filtered) {
		return []*config.GistInfo{}
	}

	return g.filtered[start:end]
}

func (g *GistManager) openSelectedGist() {
	if len(g.filtered) == 0 {
		return
	}
	gist := g.filtered[g.selected%len(g.filtered)]
	if gist.URL != "" {
		if err := openURL(gist.URL); err != nil {
			g.copyMessage = "Failed to open browser"
			g.copyMessageTime = time.Now()
		} else {
			g.copyMessage = "Opening in browser..."
			g.copyMessageTime = time.Now()
		}
	}
}

func (g *GistManager) copySelectedGistID() {
	if len(g.filtered) > 0 && g.selected < len(g.filtered) {
		gist := g.filtered[g.selected]
		clipboard.WriteAll(gist.ID)
		g.copyMessage = "Gist ID copied to clipboard!"
		g.copyMessageTime = time.Now()
	}
}

func (g *GistManager) copySelectedGistURL() {
	if len(g.filtered) > 0 && g.selected < len(g.filtered) {
		gist := g.filtered[g.selected]
		clipboard.WriteAll(gist.URL)
		g.copyMessage = "Gist URL copied to clipboard!"
		g.copyMessageTime = time.Now()
	}
}

// View renders the GistManager UI.
func (g *GistManager) View() string {
	if !g.ready {
		return "Loading..."
	}

	// Show full help if requested
	if g.showFullHelp {
		return ContainerStyle.Render(g.keys.AllKeyBindings())
	}

	var sb strings.Builder

	// Title
	sb.WriteString(TitleStyle.Render("Gist Manager") + "\n")

	// Search bar with better feedback
	if g.searchFocused {
		sb.WriteString(SearchStyle.Render(g.searchInput.View()) + "\n")
	} else {
		searchHint := "Press '/' to search"
		if g.searchTerm != "" {
			searchHint = fmt.Sprintf("Search: '%s' (Press '/' to modify)", g.searchTerm)
		}
		sb.WriteString(SubtitleStyle.Render(searchHint) + "\n")
	}

	// Filter bar
	sb.WriteString(g.renderFilterBar() + "\n\n")

	if len(g.filtered) == 0 {
		sb.WriteString(InfoStyle.Render("No gists found matching your criteria.") + "\n")
	} else {
		// Split pane layout
		leftWidth := g.width / 2
		rightWidth := g.width - leftWidth - 2
		contentHeight := g.height - 12
		leftPane := g.renderLeftPane(leftWidth, contentHeight)
		rightPane := g.renderRightPane(rightWidth, contentHeight)
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane))
	}

	if g.copyMessage != "" && time.Since(g.copyMessageTime) < 2*time.Second {
		sb.WriteString("\n" + SuccessStyle.Render(g.copyMessage))
	} else if g.copyMessage != "" && time.Since(g.copyMessageTime) >= 2*time.Second {
		g.copyMessage = ""
	}

	sb.WriteString("\n" + g.renderHelpBar())
	return ContainerStyle.Render(sb.String())
}

func (g *GistManager) renderFilterBar() string {
	var filters []string

	allStyle := ButtonStyle
	if g.filter == "all" {
		allStyle = allStyle.Copy().Background(enhancedSuccess)
	}
	filters = append(filters, allStyle.Render("All"))

	encryptedStyle := ButtonStyle
	if g.filter == "encrypted" {
		encryptedStyle = encryptedStyle.Copy().Background(enhancedSuccess)
	}
	filters = append(filters, encryptedStyle.Render("Encrypted"))

	publicStyle := ButtonStyle
	if g.filter == "public" {
		publicStyle = publicStyle.Copy().Background(enhancedSuccess)
	}
	filters = append(filters, publicStyle.Render("Public"))

	recentStyle := ButtonStyle
	if g.filter == "recent" {
		recentStyle = recentStyle.Copy().Background(enhancedSuccess)
	}
	filters = append(filters, recentStyle.Render("Recent"))

	return strings.Join(filters, " ")
}

func (g *GistManager) renderGistList(gists []*config.GistInfo, maxHeight int) string {
	var sb strings.Builder

	// Calculate the global index of the first item on this page
	pageStart := g.currentPage * g.itemsPerPage

	for i, gist := range gists {
		// Calculate global index for this item
		globalIndex := pageStart + i

		// Selection indicator
		indicator := "  "
		if globalIndex == g.selected {
			indicator = "▶ "
		}

		// Gist name and ID
		name := gist.Name
		if name == "" {
			name = gist.Description
		}
		if len(name) > 35 {
			name = name[:32] + "..."
		}

		// Status indicators
		var status []string
		if gist.IsEncrypted {
			status = append(status, "🔐")
		}
		if gist.IsPublic {
			status = append(status, "🌐")
		}
		if gist.UsageCount > 0 {
			status = append(status, fmt.Sprintf("📊%d", gist.UsageCount))
		}

		statusStr := strings.Join(status, " ")

		// Project and environment info
		var meta []string
		if gist.ProjectName != "" {
			meta = append(meta, gist.ProjectName)
		}
		if gist.Environment != "" {
			meta = append(meta, gist.Environment)
		}
		metaStr := ""
		if len(meta) > 0 {
			metaStr = " (" + strings.Join(meta, " - ") + ")"
		}

		// Format the line
		line := fmt.Sprintf("%s%s%s %s", indicator, name, metaStr, statusStr)

		// Apply selection styling
		if globalIndex == g.selected {
			line = SelectedStyle.Render(line)
		}

		sb.WriteString(line + "\n")
	}

	return sb.String()
}

func (g *GistManager) renderGistDetails(gist *config.GistInfo) string {
	var sb strings.Builder

	// Add main title
	sb.WriteString(TitleStyle.Render("Gist Details") + "\n\n")

	// Basic info section
	sb.WriteString(InfoStyle.Render("Basic Information") + "\n")
	if gist.ID != "" {
		sb.WriteString(fmt.Sprintf("ID: %s\n", gist.ID))
	}
	if gist.CreatedAt != "" {
		sb.WriteString(fmt.Sprintf("Created: %s\n", gist.CreatedAt))
	}
	if gist.UpdatedAt != "" && gist.UpdatedAt != gist.CreatedAt {
		sb.WriteString(fmt.Sprintf("Updated: %s\n", gist.UpdatedAt))
	}
	sb.WriteString(fmt.Sprintf("Files: %d\n", gist.FileCount))
	sb.WriteString(fmt.Sprintf("Usage: %d times\n", gist.UsageCount))
	if gist.LastUsed != "" {
		sb.WriteString(fmt.Sprintf("Last Used: %s\n", gist.LastUsed))
	}
	sb.WriteString("\n")

	// Status section
	sb.WriteString(InfoStyle.Render("Status") + "\n")
	if gist.IsEncrypted {
		sb.WriteString("🔐 Encrypted\n")
	} else {
		sb.WriteString("🔓 Not Encrypted\n")
	}
	if gist.IsPublic {
		sb.WriteString("🌐 Public\n")
	} else {
		sb.WriteString("🔒 Private\n")
	}
	sb.WriteString("\n")

	// Project info section (only if there's project data)
	if gist.ProjectName != "" || gist.Environment != "" {
		sb.WriteString(InfoStyle.Render("Project") + "\n")
		if gist.ProjectName != "" {
			sb.WriteString(fmt.Sprintf("Name: %s\n", gist.ProjectName))
		}
		if gist.Environment != "" {
			sb.WriteString(fmt.Sprintf("Environment: %s\n", gist.Environment))
		}
		sb.WriteString("\n")
	}

	// Tags section (only if there are tags)
	if len(gist.Tags) > 0 {
		sb.WriteString(InfoStyle.Render("Tags") + "\n")
		sb.WriteString(strings.Join(gist.Tags, ", ") + "\n")
		sb.WriteString("\n")
	}

	// URL section (ensure URL is not empty)
	if gist.URL != "" {
		sb.WriteString(InfoStyle.Render("URL") + "\n")
		sb.WriteString(gist.URL + "\n")
	}

	return sb.String()
}

func (g *GistManager) renderHelpBar() string {
	// Use the help model to render key bindings with proper wrapping
	return g.help.View(g.keys)
}

// ShowGistManager displays the interactive gist manager UI.
func ShowGistManager(gists []*config.GistInfo) error {
	model := NewGistManager(gists)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// openURL opens a URL in the default browser for the current OS.
func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser for URL %s: %w", url, err)
	}
	fmt.Printf("Opening: %s\n", url)
	return nil
}
