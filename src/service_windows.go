//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"
	"veda-anchor-engine/src/api"
	"veda-anchor-engine/src/internal/app/screentime"
	"veda-anchor-engine/src/internal/data"
	"veda-anchor-engine/src/internal/data/logger"
	"veda-anchor-engine/src/internal/ipc"
	"veda-anchor-engine/src/internal/monitoring"
	"veda-anchor-engine/src/internal/platform/nativehost"

	"golang.org/x/sys/windows/svc"
)

const serviceName = "VedaEngine"

// vedaAnchorService implements svc.Handler
type vedaAnchorService struct{}

func (s *vedaAnchorService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	// Setup logging
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "Veda", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "Veda_engine.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Println("=== VEDA ENGINE SERVICE STARTING ===")

	// Initialize core
	db, err := data.InitDB()
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		changes <- svc.Status{State: svc.StopPending}
		return true, 1
	}

	logger.NewLogger(db)
	l := logger.GetLogger()
	server := api.NewServer(db)

	// Start monitoring
	monitoring.StartDefault(l, server.Apps, screentime.StartScreenTimeMonitor)

	// Register Chrome extensions
	if err := nativehost.RegisterExtension("hkanepohpflociaodcicmmfbdaohpceo"); err != nil {
		log.Printf("Failed to register Store extension: %v", err)
	}
	if err := nativehost.RegisterExtension("gpaafgcbiejjpfdgmjglehboafdicdjb"); err != nil {
		log.Printf("Failed to register Dev extension: %v", err)
	}

	// Start IPC Server in background
	ipcServer := ipc.NewServer(server)
	go func() {
		log.Println("Veda Anchor Engine is starting IPC Server...")
		if err := ipcServer.Start(); err != nil {
			log.Printf("IPC server error: %v", err)
		}
	}()

	// Service is now running
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	log.Println("=== VEDA ENGINE SERVICE RUNNING ===")

	// Wait for stop/shutdown signal
	for {
		c := <-r
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			log.Println("=== VEDA ENGINE SERVICE STOPPING ===")
			changes <- svc.Status{State: svc.StopPending}
			// Cleanup
			l.Close()
			_ = db.Close()
			return false, 0
		case svc.Interrogate:
			changes <- c.CurrentStatus
		default:
			log.Printf("Unexpected service control request: %d", c.Cmd)
		}
	}
}

// runAsService starts the engine as a Windows Service
func runAsService() {
	err := svc.Run(serviceName, &vedaAnchorService{})
	if err != nil {
		log.Fatalf("Failed to run as service: %v", err)
	}
}
