package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger interface
type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// LoggerImpl is the implementation of Logger
type LoggerImpl struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	environment string
}

// NewLogger creates a new logger instance
func NewLogger(environment string) Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile

	return &LoggerImpl{
		infoLogger:  log.New(os.Stdout, "INFO: ", flags),
		warnLogger:  log.New(os.Stdout, "WARN: ", flags),
		errorLogger: log.New(os.Stderr, "ERROR: ", flags),
		debugLogger: log.New(os.Stdout, "DEBUG: ", flags),
		environment: environment,
	}
}

// Info logs an info message
func (l *LoggerImpl) Info(format string, args ...interface{}) {
	l.infoLogger.Output(2, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *LoggerImpl) Warn(format string, args ...interface{}) {
	l.warnLogger.Output(2, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *LoggerImpl) Error(format string, args ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal error and exits
func (l *LoggerImpl) Fatal(format string, args ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Debug logs a debug message (only in development)
func (l *LoggerImpl) Debug(format string, args ...interface{}) {
	if l.environment == "development" {
		l.debugLogger.Output(2, fmt.Sprintf(format, args...))
	}
}

// FormatTime formats a time for logging
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
