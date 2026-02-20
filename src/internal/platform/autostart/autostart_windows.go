//go:build windows

package autostart

import (
	"fmt"
	"os"
	"veda-engine/src/internal/config"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "VedaEngine"

// EnsureAutostart enables automatic startup for the VedaEngine Windows Service.
// This sets the service StartType to Automatic so it starts on boot.
func EnsureAutostart() (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to open service %q: %w", serviceName, err)
	}
	defer s.Close()

	cfg, err := s.Config()
	if err != nil {
		return "", fmt.Errorf("failed to get service config: %w", err)
	}

	cfg.StartType = mgr.StartAutomatic
	if err := s.UpdateConfig(cfg); err != nil {
		return "", fmt.Errorf("failed to update service start type: %w", err)
	}

	// Update app config to reflect the change
	appCfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		appCfg.AutostartEnabled = true
		if err := appCfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config:", err)
		}
	}

	return "", nil
}

// RemoveAutostart disables automatic startup for the VedaEngine Windows Service.
// This sets the service StartType to Disabled.
func RemoveAutostart() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		// Service doesn't exist, nothing to disable
		return nil
	}
	defer s.Close()

	cfg, err := s.Config()
	if err != nil {
		return fmt.Errorf("failed to get service config: %w", err)
	}

	cfg.StartType = mgr.StartDisabled
	if err := s.UpdateConfig(cfg); err != nil {
		return fmt.Errorf("failed to update service start type: %w", err)
	}

	// Update app config
	appCfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		appCfg.AutostartEnabled = false
		if err := appCfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config:", err)
		}
	}

	return nil
}

// GetServiceStartType queries the SCM for the current start type of the VedaEngine service.
// Returns true if the service is set to start automatically.
func GetServiceStartType() (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return false, nil // Service doesn't exist yet
	}
	defer s.Close()

	cfg, err := s.Config()
	if err != nil {
		return false, fmt.Errorf("failed to get service config: %w", err)
	}

	return cfg.StartType == mgr.StartAutomatic, nil
}

// StopAndDeleteService stops and deletes the VedaEngine service.
// Used during uninstall.
func StopAndDeleteService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return nil // Service doesn't exist
	}
	defer s.Close()

	// Try to stop the service
	_, _ = s.Control(svc.Stop)

	// Delete it
	return s.Delete()
}
