package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level   LogLevel
	debug   bool
	prefix  string
	logger  *log.Logger
	logFile *os.File
}

// NewLogger creates a new logger instance
func NewLogger(level string, debug bool, prefix string) *Logger {
	logLevel := parseLogLevel(level)

	logger := &Logger{
		level:  logLevel,
		debug:  debug,
		prefix: prefix,
	}

	// Set up logging output
	logger.setupOutput()

	return logger
}

// setupOutput sets up the logging output
func (l *Logger) setupOutput() {
	// For now, use standard output
	// In a production environment, you might want to use a proper logging library
	l.logger = log.New(os.Stdout, "", log.LstdFlags)
}

// parseLogLevel parses a string log level into LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// shouldLog determines if a message should be logged based on the current log level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// formatMessage formats a log message with timestamp, level, and prefix
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()

	if l.prefix != "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, levelStr, l.prefix, message)
	}
	return fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		message := fmt.Sprintf(format, args...)
		l.logger.Print(l.formatMessage(LogLevelDebug, message))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		message := fmt.Sprintf(format, args...)
		l.logger.Print(l.formatMessage(LogLevelInfo, message))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		message := fmt.Sprintf(format, args...)
		l.logger.Print(l.formatMessage(LogLevelWarn, message))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(LogLevelError) {
		message := fmt.Sprintf(format, args...)
		l.logger.Print(l.formatMessage(LogLevelError, message))
	}
}

// Debugf logs a debug message with formatting (alias for Debug)
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(format, args...)
}

// Infof logs an info message with formatting (alias for Info)
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

// Warnf logs a warning message with formatting (alias for Warn)
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(format, args...)
}

// Errorf logs an error message with formatting (alias for Error)
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level string) {
	l.level = parseLogLevel(level)
}

// IsDebug returns true if debug logging is enabled
func (l *Logger) IsDebug() bool {
	return l.debug
}

// WithPrefix creates a new logger with an additional prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	newPrefix := l.prefix
	if newPrefix != "" {
		newPrefix += "." + prefix
	} else {
		newPrefix = prefix
	}

	return &Logger{
		level:   l.level,
		debug:   l.debug,
		prefix:  newPrefix,
		logger:  l.logger,
		logFile: l.logFile,
	}
}
