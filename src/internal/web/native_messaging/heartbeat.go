package native_messaging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"veda-anchor-engine/src/internal/config"
)

// updateHeartbeat updates a local file with the current timestamp.
// This is used by the GUI to verify that the native messaging host (and thus the extension) is active.
func updateHeartbeat() {
	heartbeatPath, err := config.GetHeartbeatPath()
	if err != nil {
		return
	}
	// Ensure directory exists
	_ = os.MkdirAll(filepath.Dir(heartbeatPath), 0755)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(heartbeatPath, []byte(timestamp), 0644)
}
