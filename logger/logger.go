// Package logger output log
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	std = NewLogger(ErrorLevel)
)

// Level represents log level
type Level uint32

const (
	// ErrorLevel level
	ErrorLevel Level = iota
	// InfoLevel level
	InfoLevel
	// DebugLevel level
	DebugLevel
)

// Errorf output error log
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

// Infof output info log
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

// Debugf output debug log
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// SetOutput setting output method
func SetOutput(w io.Writer) {
	std.out.SetOutput(w)
}

// SetLogLevel setting loglevel
func SetLogLevel(l Level) {
	std.LogLevel = l
}

// Logger logger
type Logger struct {
	LogLevel Level
	out      *log.Logger
}

// NewLogger initialize Logger
func NewLogger(logLevel Level) *Logger {
	return &Logger{
		LogLevel: logLevel,
		out:      log.New(os.Stdout, "", 0),
	}
}

// Errorf output error log
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.LogLevel < ErrorLevel {
		return
	}
	m := fmt.Sprintf(format, args...)
	l.out.Printf("[ERROR]%s\n", m)
}

// Infof output info log
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.LogLevel < InfoLevel {
		return
	}
	m := fmt.Sprintf(format, args...)
	l.out.Printf("[INFO]%s\n", m)
}

// Debugf output debug log
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.LogLevel < DebugLevel {
		return
	}
	m := fmt.Sprintf(format, args...)
	l.out.Printf("[DEBUG]%s\n", m)
}
