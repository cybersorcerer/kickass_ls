package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	Logger    *log.Logger
	logLevel  = INFO
	levelText = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}
)

func SetLevel(level LogLevel) {
	logLevel = level
}

func parseLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func InitLogger() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create log directory in user's home
	logDir := filepath.Join(homeDir, ".local", "share", "6510lsp", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %v", logDir, err)
	}

	logFilePath := filepath.Join(logDir, "6510lsp.log")

	// Truncate existing log file on startup for clean logs
	file, err := os.OpenFile(logFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %v", logFilePath, err)
	}

	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	Info("Logger initialized at %s", logFilePath)
	return nil
}

func logf(level LogLevel, format string, v ...interface{}) {
	if level >= logLevel && Logger != nil {
		Logger.Output(3, fmt.Sprintf("[%s] %s", levelText[level], fmt.Sprintf(format, v...)))
	}
}

func Debug(format string, v ...interface{}) { logf(DEBUG, format, v...) }
func Info(format string, v ...interface{})  { logf(INFO, format, v...) }
func Warn(format string, v ...interface{})  { logf(WARN, format, v...) }
func Error(format string, v ...interface{}) { logf(ERROR, format, v...) }
