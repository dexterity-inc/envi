package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dexterity-inc/envi/internal/tui"
)

// Buffer pool for efficient memory reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, DefaultBufferSize)
	},
}

// getBuffer retrieves a buffer from the pool
func getBuffer() []byte {
	return bufferPool.Get().([]byte)[:0]
}

// putBuffer returns a buffer to the pool
func putBuffer(buf []byte) {
	if cap(buf) <= DefaultBufferSize*2 { // Don't keep oversized buffers
		bufferPool.Put(buf)
	}
}

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger provides centralized logging functionality
type Logger struct {
	useTUI bool
	level  LogLevel
}

// NewLogger creates a new logger instance
func NewLogger(useTUI bool) *Logger {
	return &Logger{
		useTUI: useTUI,
		level:  INFO,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		message := fmt.Sprintf(format, args...)
		if l.useTUI {
			// In TUI mode, debug messages are typically not shown
			// but could be logged to a file if needed
		} else {
			fmt.Printf("[DEBUG] %s\n", message)
		}
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		message := fmt.Sprintf(format, args...)
		if l.useTUI {
			tui.DisplayMessage("Info", message)
		} else {
			fmt.Printf("ℹ️  %s\n", message)
		}
	}
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.useTUI {
		tui.DisplayMessage("Success", message)
	} else {
		fmt.Printf("✅ %s\n", message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		message := fmt.Sprintf(format, args...)
		if l.useTUI {
			tui.DisplayMessage("Warning", message)
		} else {
			fmt.Printf("⚠️  %s\n", message)
		}
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		message := fmt.Sprintf(format, args...)
		if l.useTUI {
			tui.DisplayMessage("Error", message)
		} else {
			fmt.Printf("❌ %s\n", message)
		}
	}
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.useTUI {
		tui.DisplayMessage("Fatal Error", message)
	} else {
		fmt.Printf("💥 %s\n", message)
	}
	os.Exit(1)
}

// ErrorWithContext logs an error with additional context
func (l *Logger) ErrorWithContext(err error, context string, suggestions ...string) {
	if l.useTUI {
		message := fmt.Sprintf("Error in %s: %s", context, err)
		if len(suggestions) > 0 {
			message += "\n\nSuggestions:\n"
			for _, suggestion := range suggestions {
				message += fmt.Sprintf("• %s\n", suggestion)
			}
		}
		tui.DisplayMessage("Error", message)
	} else {
		fmt.Printf("❌ Error in %s: %s\n", context, err)
		if len(suggestions) > 0 {
			fmt.Println("💡 Suggestions:")
			for _, suggestion := range suggestions {
				fmt.Printf("   • %s\n", suggestion)
			}
		}
	}
}

// Confirm prompts for user confirmation
func (l *Logger) Confirm(title, question string) (bool, error) {
	if l.useTUI {
		return tui.Confirm(title, question)
	} else {
		fmt.Printf("%s (y/N): ", question)
		var response string
		fmt.Scanln(&response)
		return strings.ToLower(response) == "y", nil
	}
}

// GetPassword prompts for a password
func (l *Logger) GetPassword(title string, confirm bool) (string, error) {
	if l.useTUI {
		return tui.GetPassword(title, confirm)
	} else {
		fmt.Print("Enter password: ")
		// Note: This is a simplified version. In production, you'd want to use
		// golang.org/x/term for secure password input
		var password string
		fmt.Scanln(&password)
		return password, nil
	}
}

// GetText prompts for text input
func (l *Logger) GetText(title, label, description, placeholder, help string, required bool) (string, error) {
	if l.useTUI {
		return tui.GetText(title, label, description, placeholder, help, required)
	} else {
		fmt.Printf("%s: ", label)
		if description != "" {
			fmt.Printf("(%s) ", description)
		}
		var text string
		fmt.Scanln(&text)
		return text, nil
	}
}

// Progress shows progress information
func (l *Logger) Progress(message string, percent float64) {
	// For now, just use console output since the TUI progress function doesn't exist
	fmt.Printf("🔄 %s: %.1f%%\n", message, percent*100)
}

// ProgressComplete shows completion of a progress operation
func (l *Logger) ProgressComplete(message string) {
	// For now, just use console output since the TUI progress function doesn't exist
	fmt.Printf("✅ %s\n", message)
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger
func InitLogger(useTUI bool) {
	globalLogger = NewLogger(useTUI)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewLogger(true) // Default to TUI mode
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Success(format string, args ...interface{}) {
	GetLogger().Success(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

func ErrorWithContext(err error, context string, suggestions ...string) {
	GetLogger().ErrorWithContext(err, context, suggestions...)
}

func Confirm(title, question string) (bool, error) {
	return GetLogger().Confirm(title, question)
}

func GetPassword(title string, confirm bool) (string, error) {
	return GetLogger().GetPassword(title, confirm)
}

func GetText(title, label, description, placeholder, help string, required bool) (string, error) {
	return GetLogger().GetText(title, label, description, placeholder, help, required)
}

func Progress(message string, percent float64) {
	GetLogger().Progress(message, percent)
}

func ProgressComplete(message string) {
	GetLogger().ProgressComplete(message)
}
