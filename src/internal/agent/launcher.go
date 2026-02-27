//go:build windows

package agent

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

const (
	INVALID_SESSION_ID = ^uint32(0)
)

func LaunchAgent() error {
	agentPath := filepath.Join(os.Getenv("ProgramFiles"), "VedaAnchor", "veda-anchor-agent.exe")

	if _, err := os.Stat(agentPath); err != nil {
		log.Printf("[AgentLauncher] Agent binary not found: %s", agentPath)
		return err
	}

	sessionID, err := getActiveConsoleSession()
	if err != nil {
		log.Printf("[AgentLauncher] Failed to get active session: %v", err)
		return err
	}

	if sessionID == INVALID_SESSION_ID {
		log.Printf("[AgentLauncher] No active user session")
		return nil
	}

	log.Printf("[AgentLauncher] Found active session: %d", sessionID)

	token, err := getUserToken(sessionID)
	if err != nil {
		log.Printf("[AgentLauncher] Failed to get user token: %v", err)
		return err
	}
	defer token.Close()

	log.Printf("[AgentLauncher] Launching Agent in user session...")
	err = createProcessAsUser(token, agentPath)
	if err != nil {
		log.Printf("[AgentLauncher] Failed to launch Agent: %v", err)
		return err
	}

	log.Printf("[AgentLauncher] Agent launched successfully")
	return nil
}

func getActiveConsoleSession() (uint32, error) {
	sessionID := windows.WTSGetActiveConsoleSessionId()
	if sessionID == INVALID_SESSION_ID {
		return INVALID_SESSION_ID, nil
	}
	return sessionID, nil
}

func getUserToken(sessionID uint32) (windows.Token, error) {
	var token windows.Token
	err := windows.WTSQueryUserToken(sessionID, &token)
	if err != nil {
		return 0, err
	}

	var tokenDup windows.Token
	err = windows.DuplicateTokenEx(
		token,
		windows.MAXIMUM_ALLOWED,
		nil,
		windows.SecurityIdentification,
		windows.TokenPrimary,
		&tokenDup,
	)
	token.Close()
	if err != nil {
		return 0, err
	}

	return tokenDup, nil
}

func createProcessAsUser(token windows.Token, exePath string) error {
	appName, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return err
	}

	si := windows.StartupInfo{
		Desktop:    windows.StringToUTF16Ptr("Winsta0\\default"),
		Flags:      windows.STARTF_USESHOWWINDOW,
		ShowWindow: windows.SW_HIDE,
	}
	var pi windows.ProcessInformation

	err = windows.CreateProcessAsUser(
		token,
		appName,
		nil,
		nil,
		nil,
		false,
		0,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		return err
	}

	windows.CloseHandle(pi.Process)
	windows.CloseHandle(pi.Thread)
	return nil
}

func StartAgentWithRetry() {
	for i := 0; i < 10; i++ {
		time.Sleep(2 * time.Second)
		if err := LaunchAgent(); err == nil {
			return
		}
	}
	log.Printf("[AgentLauncher] Giving up after 10 attempts")
}
