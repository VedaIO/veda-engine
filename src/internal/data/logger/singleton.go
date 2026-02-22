package logger

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"
	"veda-anchor-engine/src/internal/config"
)

var defaultLogger Logger
var once sync.Once

// NewLogger initializes the default singleton logger.
// It uses a sync.Once to ensure that the logger is only initialized once, making it safe for concurrent use.
func NewLogger(db *sql.DB) {
	once.Do(func() {
		logPath, err := config.GetLogPath()
		if err != nil {
			log.Fatalf("Failed to get log path: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		fileLogger := log.New(file, "", log.LstdFlags)
		defaultLogger = &multiLogger{db: db, file: file, logger: fileLogger, mu: sync.Mutex{}}
	})
}

// GetLogger returns the default singleton logger instance.
func GetLogger() Logger {
	return defaultLogger
}
