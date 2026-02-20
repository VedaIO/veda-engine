//go:generate goversioninfo -64

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"veda-engine/src/api"
	"veda-engine/src/internal/app/screentime"
	"veda-engine/src/internal/data"
	"veda-engine/src/internal/data/logger"
	"veda-engine/src/internal/ipc"
	"veda-engine/src/internal/monitoring"
	"veda-engine/src/internal/platform/nativehost"
	"veda-engine/src/internal/web/native_messaging"

	"golang.org/x/sys/windows/svc"
)

func main() {
	// Check if running as a Windows Service (started by SCM)
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}

	if isService {
		runAsService()
		return
	}

	// --- INTERACTIVE MODE (debug/dev or native messaging) ---

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

	log.Printf("=== VEDA ENGINE LAUNCHED (interactive) === Args: %v", os.Args)

	// NATIVE MESSAGING HOST MODE
	// Chrome launches us with the extension ID as an argument: chrome-extension://...
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		log.Println("[MODE] Native Messaging Host detected")
		native_messaging.Run()
		log.Println("[MODE] Native Messaging Host exited")
		return
	}

	// STANDALONE INTERACTIVE MODE (for debugging)
	log.Println("[MODE] Interactive Standalone")

	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	logger.NewLogger(db)
	l := logger.GetLogger()
	server := api.NewServer(db)

	// Start background monitoring
	monitoring.StartDefault(l, server.Apps, screentime.StartScreenTimeMonitor)

	// Register Chrome extensions
	if err := nativehost.RegisterExtension("hkanepohpflociaodcicmmfbdaohpceo"); err != nil {
		log.Printf("Failed to register Store extension: %v", err)
	}
	if err := nativehost.RegisterExtension("gpaafgcbiejjpfdgmjglehboafdicdjb"); err != nil {
		log.Printf("Failed to register Dev extension: %v", err)
	}

	// Start IPC Server (blocking)
	ipcServer := ipc.NewServer(server)
	log.Println("Veda Engine is starting IPC Server...")
	if err := ipcServer.Start(); err != nil {
		log.Fatalf("Failed to start IPC server: %v", err)
	}

	select {}
}
