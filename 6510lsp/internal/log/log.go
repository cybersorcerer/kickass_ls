package log

import (
	"log"
	"os"
	"path/filepath"
)

var Logger *log.Logger

func InitLogger() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".local", "share", "6510lsp", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFilePath := filepath.Join(logDir, "6510lsp.log")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	Logger = log.New(file, "6510lsp: ", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Logger initialized.")
	return nil
}
