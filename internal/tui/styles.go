package tui

import "github.com/charmbracelet/lipgloss"

// Enhanced color scheme
var (
	enhancedPrimary = lipgloss.AdaptiveColor{Light: "#1a73e8", Dark: "#4285f4"}
	enhancedSuccess = lipgloss.AdaptiveColor{Light: "#34a853", Dark: "#0f9d58"}
	enhancedError   = lipgloss.AdaptiveColor{Light: "#ea4335", Dark: "#f28b82"}
	enhancedWarning = lipgloss.AdaptiveColor{Light: "#fbbc04", Dark: "#fdd663"}
	enhancedInfo    = lipgloss.AdaptiveColor{Light: "#4285f4", Dark: "#8ab4f8"}
	enhancedText    = lipgloss.AdaptiveColor{Light: "#202124", Dark: "#e8eaed"}
	enhancedSubtext = lipgloss.AdaptiveColor{Light: "#5f6368", Dark: "#9aa0a6"}
	enhancedBorder  = lipgloss.AdaptiveColor{Light: "#dadce0", Dark: "#5f6368"}
	enhancedBg      = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#202124"}
	enhancedAccent  = lipgloss.AdaptiveColor{Light: "#f1f3f4", Dark: "#3c4043"}
)

// Enhanced styles with better responsive design
var (
	// ContainerStyle provides the main container styling
	ContainerStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(enhancedBorder).
			Background(enhancedBg)

	// TitleStyle for main titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(enhancedPrimary).
			MarginBottom(1).
			Align(lipgloss.Center)

	// SubtitleStyle for secondary titles
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(enhancedSubtext).
			Italic(true).
			MarginBottom(2)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(enhancedSuccess).
			Bold(true)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(enhancedError).
			Bold(true)

	// WarningStyle for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(enhancedWarning).
			Bold(true)

	// InfoStyle for informational messages
	InfoStyle = lipgloss.NewStyle().
			Foreground(enhancedInfo).
			Bold(true)

	// ButtonStyle for interactive buttons
	ButtonStyle = lipgloss.NewStyle().
			Foreground(enhancedText).
			Background(enhancedPrimary).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true)

	// SelectedStyle for selected items
	SelectedStyle = lipgloss.NewStyle().
			Foreground(enhancedPrimary).
			Bold(true).
			Background(enhancedAccent)

	// SearchStyle for search input
	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(enhancedPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// HelpStyle for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(enhancedSubtext).
			Italic(true).
			MarginTop(1)
)
