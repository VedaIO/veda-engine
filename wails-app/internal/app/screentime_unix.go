//go:build !windows

package app

import (
	"database/sql"
	"wails-app/internal/data"
)

// StartScreenTimeMonitor is a no-op on non-Windows platforms
func StartScreenTimeMonitor(appLogger data.Logger, db *sql.DB) {
	// Screen time monitoring is only supported on Windows
}
