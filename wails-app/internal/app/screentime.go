//go:build windows

package app

import (
	"database/sql"
	"strings"
	"time"
	"wails-app/internal/data"
	"wails-app/internal/native/screentime"

	"github.com/shirou/gopsutil/v3/process"
)

const screenTimeCheckInterval = 1 * time.Second

// ScreenTimeState tracks the current foreground window state
type ScreenTimeState struct {
	lastExePath string
	lastTitle   string
	lastPID     uint32
}

// StartScreenTimeMonitor starts a goroutine that tracks foreground window time
func StartScreenTimeMonitor(appLogger data.Logger, db *sql.DB) {
	go func() {
		state := &ScreenTimeState{}
		ticker := time.NewTicker(screenTimeCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			trackForegroundWindow(appLogger, state)
		}
	}()
}

// trackForegroundWindow checks the current foreground window and logs screen time
func trackForegroundWindow(appLogger data.Logger, state *ScreenTimeState) {
	// Get foreground window info from CGO
	info := screentime.GetActiveWindowInfo()
	if info == nil || info.PID == 0 {
		return
	}

	// Get executable path from PID
	proc, err := process.NewProcess(int32(info.PID))
	if err != nil {
		return
	}

	exePath, err := proc.Exe()
	if err != nil {
		return
	}

	// Skip ProcGuard itself
	if strings.Contains(strings.ToLower(exePath), "procguard.exe") {
		return
	}

	// Get the current timestamp (start of day for grouping)
	now := time.Now().Unix()

	// Check if this is the same window as before
	if exePath == state.lastExePath && info.Title == state.lastTitle {
		// Same window - update the most recent record by incrementing duration
		data.EnqueueWrite(`
			UPDATE screen_time 
			SET duration_seconds = duration_seconds + 1
			WHERE id = (
				SELECT id FROM screen_time 
				WHERE executable_path = ? AND window_title = ?
				ORDER BY timestamp DESC LIMIT 1
			)
		`, exePath, info.Title)
	} else {
		// Different window - insert new record
		data.EnqueueWrite(`
			INSERT INTO screen_time (executable_path, window_title, timestamp, duration_seconds)
			VALUES (?, ?, ?, 1)
		`, exePath, info.Title, now)

		// Update state
		state.lastExePath = exePath
		state.lastTitle = info.Title
		state.lastPID = info.PID
	}
}
