package monitoring

import (
	"os"
	"slices"
	"strings"

	"src/internal/blocklist/app"
	"src/internal/data/logger"
)

type BlocklistSubscriber struct {
	logger logger.Logger
}

func NewBlocklistSubscriber(appLogger logger.Logger) *BlocklistSubscriber {
	return &BlocklistSubscriber{
		logger: appLogger,
	}
}

func (s *BlocklistSubscriber) Name() string {
	return "BlocklistSubscriber"
}

func (s *BlocklistSubscriber) OnProcessesChanged(snapshot ProcessSnapshot) {
	blocklist, err := app.LoadAppBlocklist()
	if err != nil {
		s.logger.Printf("[BlocklistSubscriber] Failed to load blocklist: %v", err)
		return
	}

	if len(blocklist) == 0 {
		return
	}

	for _, proc := range snapshot.Processes {
		procName := proc.Name
		if procName == "" {
			continue
		}

		if slices.Contains(blocklist, strings.ToLower(procName)) {
			osProc, err := os.FindProcess(int(proc.PID))
			if err == nil {
				if err := osProc.Kill(); err != nil {
					s.logger.Printf("[BlocklistSubscriber] Failed to kill process %s (pid %d): %v", procName, proc.PID, err)
				} else {
					s.logger.Printf("[BlocklistSubscriber] Killed blocked process %s (pid %d)", procName, proc.PID)
				}
			}
		}
	}
}
