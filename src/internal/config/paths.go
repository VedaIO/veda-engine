package config

import (
	"os"
	"path/filepath"
)

// GetAppRoot returns the root directory for application data and logs.
// We use ProgramData for shared access between the Windows Service (System) and UI (User).
func GetAppRoot() (string, error) {
	progData := os.Getenv("ProgramData")
	if progData == "" {
		progData = `C:\ProgramData`
	}
	return filepath.Join(progData, AppName), nil
}

// GetLogDir returns the directory where log files are stored.
func GetLogDir() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "logs"), nil
}

// GetLogPath returns the full path to the engine log file.
func GetLogPath() (string, error) {
	dir, err := GetLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "veda-anchor_engine.log"), nil
}

// GetNativeHostLogPath returns the full path to the native messaging host log file.
func GetNativeHostLogPath() (string, error) {
	dir, err := GetLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "native_host.log"), nil
}

// GetDataDir returns the directory where data files (database, etc.) are stored.
func GetDataDir() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "data"), nil
}

// GetConfigDir returns the directory where configuration files are stored.
func GetConfigDir() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config"), nil
}

// GetDatabasePath returns the full path to the SQLite database file.
func GetDatabasePath() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "veda-anchor.db"), nil
}

// GetHeartbeatPath returns the path to the extension heartbeat file.
func GetHeartbeatPath() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "extension_heartbeat"), nil
}

// GetNativeHostManifestPath returns the full path to the native messaging host manifest file.
func GetNativeHostManifestPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "native-host.json"), nil
}

// GetWebBlocklistPath returns the full path to the web blocklist file.
func GetWebBlocklistPath() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "veda-anchor_web_blocklist.json"), nil
}

// GetAppBlocklistPath returns the full path to the app blocklist file.
func GetAppBlocklistPath() (string, error) {
	root, err := GetAppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "veda-anchor_app_blocklist.json"), nil
}
