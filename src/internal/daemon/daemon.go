package daemon

import (
	"time"

	"src/internal/app/screentime"
	"src/internal/data/logger"
	"src/internal/data/repository"
	"src/internal/monitoring"
	"src/internal/platform/autostart"
)

const monitoringPollingInterval = 2 * time.Second

var resetCh = make(chan struct{}, 1)

func Start(appLogger logger.Logger, apps *repository.AppRepository, web *repository.WebRepository) {
	if _, err := autostart.EnsureAutostart(); err != nil {
		appLogger.Printf("Failed to set up autostart: %v", err)
	}

	manager := monitoring.NewMonitoringManager(appLogger, monitoringPollingInterval)

	processEventSubscriber := monitoring.NewProcessEventSubscriber(appLogger, apps)
	processEventSubscriber.InitializeFromDatabase()
	manager.RegisterSubscriber(processEventSubscriber)

	blocklistSubscriber := monitoring.NewBlocklistSubscriber(appLogger)
	manager.RegisterSubscriber(blocklistSubscriber)

	monitoring.SetGlobalManager(manager)

	manager.Start()

	screentime.StartScreenTimeMonitor(appLogger, apps, web)
}

func ResetMonitoring() {
	resetCh <- struct{}{}
}

func WaitForResetSignal() <-chan struct{} {
	return resetCh
}
