package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Color scheme - soft colors that work well in both light and dark terminals
var (
	primaryColor   = lipgloss.AdaptiveColor{Light: "#2975c0", Dark: "#4b95d6"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#28a745", Dark: "#3ebd5a"}
	errorColor     = lipgloss.AdaptiveColor{Light: "#dc3545", Dark: "#e35d6a"}
	warningColor   = lipgloss.AdaptiveColor{Light: "#ffc107", Dark: "#ffcf33"}
	textColor      = lipgloss.AdaptiveColor{Light: "#212529", Dark: "#f8f9fa"}
	subtextColor   = lipgloss.AdaptiveColor{Light: "#6c757d", Dark: "#adb5bd"}
	borderColor    = lipgloss.AdaptiveColor{Light: "#4b95d6", Dark: "#4b95d6"}
	bgColor        = lipgloss.AdaptiveColor{Light: "#f8f9fa", Dark: "#343a40"}
	highlightColor = lipgloss.AdaptiveColor{Light: "#e9ecef", Dark: "#495057"}
)

// UI style definitions 
var (
	// Container style
	appStyle = lipgloss.NewStyle().
		MarginLeft(0).
		MarginRight(0)

	// Title style
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(0, 0).
		MarginBottom(1).
		Align(lipgloss.Center)

	// Description style
	descriptionStyle = lipgloss.NewStyle().
		Foreground(subtextColor).
		Align(lipgloss.Center).
		MarginBottom(1)

	// Field container style
	fieldContainerStyle = lipgloss.NewStyle().
		MarginBottom(1)

	// Focused input style
	focusedInputStyle = lipgloss.NewStyle().
		Foreground(textColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1)

	// Blurred input style
	blurredInputStyle = lipgloss.NewStyle().
		Foreground(subtextColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(subtextColor).
		Padding(0, 1)

	// Label style
	labelStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		MarginBottom(0)

	// Help text style
	helpTextStyle = lipgloss.NewStyle().
		Foreground(subtextColor).
		Italic(true).
		PaddingLeft(2).
		MarginTop(0)

	// Error message style
	errorMessageStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		MarginBottom(1)

	// Success style
	successStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true)

	// Help bar style
	helpStyle = lipgloss.NewStyle().
		Foreground(subtextColor).
		MarginTop(1)

	// Required field marker
	requiredStyle = lipgloss.NewStyle().
		Foreground(errorColor)

	// Button style
	buttonStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Background(primaryColor).
		Padding(0, 1).
		Bold(true).
		MarginTop(1)
)

// InputField represents a field that requires user input
type InputField struct {
	Label       string // Display label
	Placeholder string // Placeholder text
	Value       string // Default value
	Secret      bool   // Whether to mask input (for passwords)
	Required    bool   // Whether the field is required
	Help        string // Help text for the field
}

// InputModel manages the input form state
type InputModel struct {
	title       string
	description string
	inputs      []textinput.Model
	fields      []InputField
	focusIndex  int
	submitted   bool
	err         error
	help        help.Model
	keyMap      keyMap
	width       int
	height      int
	ready       bool
	showHelp    bool
}

// keyMap defines keyboard shortcuts
type keyMap struct {
	Next     key.Binding
	Prev     key.Binding
	Submit   key.Binding
	Quit     key.Binding
	Help     key.Binding
	ShowHelp key.Binding
}

// ShortHelp returns keybindings to show in the short help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Submit, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Submit},
		{k.Help, k.ShowHelp, k.Quit},
	}
}

// New creates a new input form model
func New(title, description string, fields []InputField) InputModel {
	inputs := make([]textinput.Model, len(fields))
	
	// Set up each input field
	for i, field := range fields {
		t := textinput.New()
		t.Placeholder = field.Placeholder
		t.Width = 25 
		t.Prompt = ""
		
		// Configure password masking if needed
		if field.Secret {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		
		// Set initial value if provided
		if field.Value != "" {
			t.SetValue(field.Value)
		}
		
		// Focus first field by default
		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
			t.TextStyle = lipgloss.NewStyle().Foreground(textColor)
		}
		
		inputs[i] = t
	}

	// Set up keyboard shortcuts
	keymap := keyMap{
		Next: key.NewBinding(
			key.WithKeys("tab", "down", "ctrl+n"),
			key.WithHelp("tab", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab", "up", "ctrl+p"),
			key.WithHelp("shift+tab", "prev"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "help"),
		),
		ShowHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}

	return InputModel{
		title:       title,
		description: description,
		inputs:      inputs,
		fields:      fields,
		keyMap:      keymap,
		help:        help.New(),
		showHelp:    false,
	}
}

// Start runs the input form and returns the entered values
func (m InputModel) Start() (map[string]string, error) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	model, err := p.Run()
	if err != nil {
		return nil, err
	}
	
	finalModel := model.(InputModel)
	
	// Check if form was submitted or canceled
	if !finalModel.submitted {
		return nil, fmt.Errorf("canceled")
	}
	
	// Collect input values
	values := make(map[string]string)
	for i, input := range finalModel.inputs {
		values[finalModel.fields[i].Label] = input.Value()
	}
	
	return values, nil
}

// Init initializes the model
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input and updates the model state
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
			
		case key.Matches(msg, m.keyMap.Next):
			m.nextInput()
			
		case key.Matches(msg, m.keyMap.Prev):
			m.prevInput()
			
		case key.Matches(msg, m.keyMap.Submit):
			if m.validateFields() {
				m.submitted = true
				return m, tea.Quit
			}
			
		case key.Matches(msg, m.keyMap.ShowHelp):
			m.showHelp = !m.showHelp
			
		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
		}
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		
		// Adjust the help view
		m.help.Width = msg.Width
	}
	
	// Handle input updates
	cmd := m.updateInputs(msg)
	
	return m, cmd
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// nextInput focuses the next input field
func (m *InputModel) nextInput() {
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	m.updateFocus()
}

// prevInput focuses the previous input field
func (m *InputModel) prevInput() {
	m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
	m.updateFocus()
}

// updateFocus updates the focus state of all inputs
func (m *InputModel) updateFocus() {
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIndex {
			// Focus the selected field
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
			m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(textColor)
		} else {
			// Blur all other fields
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(subtextColor)
			m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(subtextColor)
		}
	}
}

// validateFields checks if all required fields have values
func (m *InputModel) validateFields() bool {
	for i, field := range m.fields {
		if field.Required && m.inputs[i].Value() == "" {
			m.err = fmt.Errorf("required field '%s' cannot be empty", field.Label)
			// Focus the empty required field
			m.focusIndex = i
			m.updateFocus()
			return false
		}
	}
	m.err = nil
	return true
}

// updateInputs sends the update message to the focused input only
func (m *InputModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	
	// Only update the focused input
	for i := range m.inputs {
		if i == m.focusIndex {
			var cmd tea.Cmd
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	
	return tea.Batch(cmds...)
}

// renderField formats a single input field with its label and help text
func (m *InputModel) renderField(index int) string {
	field := m.fields[index]
	input := m.inputs[index]
	
	// Determine if this field is focused
	var style lipgloss.Style
	if index == m.focusIndex {
		style = focusedInputStyle
	} else {
		style = blurredInputStyle
	}
	
	// Create the input field with appropriate styling
	renderedInput := style.Render(input.View())
	
	// Add label with required indicator if needed
	label := labelStyle.Render(field.Label)
	if field.Required {
		label += " " + requiredStyle.Render("*")
	}
	
	// Add help text if provided
	var helpText string
	if field.Help != "" {
		helpText = helpTextStyle.Render(field.Help)
	}
	
	// Combine all elements
	return fmt.Sprintf("%s\n%s\n%s", label, renderedInput, helpText)
}

// View renders the UI
func (m InputModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	
	// Build the form view
	var b strings.Builder
	
	// Title and description
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n")
	
	if m.description != "" {
		b.WriteString(descriptionStyle.Render(m.description))
		b.WriteString("\n")
	}
	
	// Error message if any
	if m.err != nil {
		b.WriteString(errorMessageStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}
	
	// Render each input field
	for i := range m.inputs {
		b.WriteString(fieldContainerStyle.Render(m.renderField(i)))
		b.WriteString("\n")
	}
	
	// Add instruction text
	b.WriteString("\n")
	if m.showHelp {
		b.WriteString(m.help.View(m.keyMap))
	} else {
		b.WriteString(helpStyle.Render("Press ? for help, tab to navigate, enter to submit"))
	}
	
	// Center the content
	return appStyle.Render(b.String())
}

// GetPassword prompts for a password with optional confirmation
func GetPassword(title string, confirm bool) (string, error) {
	// Set up the password field
	var fields []InputField
	
	fields = append(fields, InputField{
		Label:       "Password",
		Placeholder: "Enter password",
		Secret:      true,
		Required:    true,
		Help:        "Minimum 8 characters",
	})
	
	// Add confirmation field if requested
	if confirm {
		fields = append(fields, InputField{
			Label:       "Confirm",
			Placeholder: "Confirm password",
			Secret:      true,
			Required:    true,
			Help:        "Re-enter the same password",
		})
	}
	
	// Create the input model
	model := New(title, "Password will not be displayed", fields)
	
	// Run the form
	result, err := model.Start()
	if err != nil {
		return "", err
	}
	
	// For confirmation, check that passwords match
	if confirm {
		password := result["Password"]
		confirmation := result["Confirm"]
		
		if password != confirmation {
			return "", fmt.Errorf("passwords do not match")
		}
	}
	
	return result["Password"], nil
}

// GetText prompts for a text input with the given parameters
func GetText(title, label, description, placeholder, help string, required bool) (string, error) {
	fields := []InputField{
		{
			Label:       label,
			Placeholder: placeholder,
			Help:        help,
			Required:    required,
		},
	}
	
	model := New(title, description, fields)
	
	result, err := model.Start()
	if err != nil {
		return "", err
	}
	
	return result[label], nil
}

// GetConfirmation prompts for a yes/no confirmation
func GetConfirmation(title, question string) (bool, error) {
	fields := []InputField{
		{
			Label:       "Confirm",
			Placeholder: "yes/no",
			Help:        "Type 'yes' or 'no'",
			Required:    true,
		},
	}
	
	model := New(title, question, fields)
	
	result, err := model.Start()
	if err != nil {
		return false, err
	}
	
	answer := strings.ToLower(strings.TrimSpace(result["Confirm"]))
	
	return answer == "yes" || answer == "y", nil
}

// DisplayMessage shows a message with a scrollable viewport
func DisplayMessage(title, message string) error {
	m := viewportModel{
		title:   title,
		content: message,
	}
	
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	_, err := p.Run()
	return err
}

// viewportModel is a simple model for displaying scrollable content
type viewportModel struct {
	title    string
	content  string
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

// Init initializes the viewport model
func (m viewportModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the viewport state
func (m viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		if !m.ready {
			// Initialize viewport with dimensions
			headerHeight := 3 // title + padding
			footerHeight := 2 // help text
			
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.SetContent(m.content)
			m.viewport.Style = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(borderColor)
			
			// Enable scrollbar if content is long enough
			if strings.Count(m.content, "\n") > m.viewport.Height {
				m.viewport.YPosition = 0
				m.viewport.HighPerformanceRendering = true
			}
			
			m.ready = true
		} else {
			// Resize the viewport
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 5
		}
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
			
		case "up", "k":
			m.viewport.LineUp(1)
			
		case "down", "j":
			m.viewport.LineDown(1)
			
		case "pgup":
			m.viewport.LineUp(10)
			
		case "pgdown":
			m.viewport.LineDown(10)
			
		case "home":
			m.viewport.GotoTop()
			
		case "end":
			m.viewport.GotoBottom()
		}
	}
	
	// Update viewport and collect commands
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m viewportModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	
	var sb strings.Builder
	
	// Add title
	sb.WriteString(titleStyle.Render(m.title))
	sb.WriteString("\n\n")
	
	// Add viewport with content
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n\n")
	
	// Add help text
	sb.WriteString(helpStyle.Render("Press q or esc to exit, ↑/↓ to scroll"))
	
	return sb.String()
}

// Confirm displays a confirmation prompt
func Confirm(title, message string) (bool, error) {
	return GetConfirmation(title, message)
} 