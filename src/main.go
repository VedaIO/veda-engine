//go:generate goversioninfo -64

package main

import (
	"log"
	"os"
	"path/filepath"
	"veda-engine/src/api"
	"veda-engine/src/internal/app/screentime"
	"veda-engine/src/internal/data"
	"veda-engine/src/internal/data/logger"
	"veda-engine/src/internal/ipc"
	"veda-engine/src/internal/monitoring"
	"veda-engine/src/internal/platform/nativehost"
	"veda-engine/src/internal/web/native_messaging"
	"strings"
)

func main() {
	// CRITICAL: Log startup for debugging
	// Use absolute path in CacheDir because CWD varies when launched by Chrome
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "Veda", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "Veda_engine.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Printf("=== VEDA ENGINE LAUNCHED === Args: %v", os.Args)
	log.Printf("CWD: %v", func() string { wd, _ := os.Getwd(); return wd }())

	// MODE 1: NATIVE MESSAGING HOST
	// Chrome launches us with the extension ID as an argument: chrome-extension://...
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		log.Println("[MODE] Native Messaging Host detected")
		native_messaging.Run()
		log.Println("[MODE] Native Messaging Host exited")
		return
	}

	// MODE 2: STANDALONE SERVICE
	log.Println("[MODE] Standalone Service detected")

	// Initialize database connection
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize logger with database
	logger.NewLogger(db)
	l := logger.GetLogger()

	// Create API server with database connection
	server := api.NewServer(db)

	// Start background monitoring services.
	monitoring.StartDefault(l, server.Apps, screentime.StartScreenTimeMonitor)

	// Ensure Native Messaging Host is registered
	if err := nativehost.RegisterExtension("hkanepohpflociaodcicmmfbdaohpceo"); err != nil {
		log.Printf("Failed to register Store extension: %v", err)
	}
	if err := nativehost.RegisterExtension("gpaafgcbiejjpfdgmjglehboafdicdjb"); err != nil {
		log.Printf("Failed to register Dev extension: %v", err)
	}

	// Start IPC Server
	ipcServer := ipc.NewServer(server)
	log.Println("Veda Engine is starting IPC Server...")
	if err := ipcServer.Start(); err != nil {
		log.Fatalf("Failed to start IPC server: %v", err)
	}

	// Keep the service running (actually listener.Accept() in Start() is blocking,
	// but if Start() were async we'd need this select.
	// Current Start() is blocking so we won't even reach here until it stops.)
	select {}
}
