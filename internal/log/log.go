package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Logger represents a logger for a process
type Logger struct {
	processID   string
	logDir      string
	stdoutFile  *os.File
	stderrFile  *os.File
}

// NewLogger creates a new logger for a process
func NewLogger(processID string) (*Logger, error) {
	logDir := filepath.Join(os.TempDir(), "pm3", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	stdoutPath := filepath.Join(logDir, fmt.Sprintf("%s-out.log", processID))
	stderrPath := filepath.Join(logDir, fmt.Sprintf("%s-err.log", processID))

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stdoutFile.Close()
		return nil, err
	}

	return &Logger{
		processID:   processID,
		logDir:      logDir,
		stdoutFile:  stdoutFile,
		stderrFile:  stderrFile,
	}, nil
}

// Stdout returns the stdout writer
func (l *Logger) Stdout() io.Writer {
	return io.MultiWriter(os.Stdout, l.stdoutFile)
}

// Stderr returns the stderr writer
func (l *Logger) Stderr() io.Writer {
	return io.MultiWriter(os.Stderr, l.stderrFile)
}

// Close closes the logger
func (l *Logger) Close() error {
	if err := l.stdoutFile.Close(); err != nil {
		return err
	}
	return l.stderrFile.Close()
}

// GetLogPath returns the path to the log files
func (l *Logger) GetLogPath() string {
	return l.logDir
}

// ReadLogs reads logs from the log files
func (l *Logger) ReadLogs() (string, error) {
	stdout, err := os.ReadFile(filepath.Join(l.logDir, fmt.Sprintf("%s-out.log", l.processID)))
	if err != nil {
		return "", err
	}

	stderr, err := os.ReadFile(filepath.Join(l.logDir, fmt.Sprintf("%s-err.log", l.processID)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("=== STDOUT ===\n%s\n=== STDERR ===\n%s", stdout, stderr), nil
}
