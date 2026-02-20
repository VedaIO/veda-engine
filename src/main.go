//go:generate goversioninfo -64

package main

import (
	"log"

	"golang.org/x/sys/windows/svc"
)

func main() {
	// veda-engine runs exclusively as a Windows Service.
	// It is registered and started by the veda launcher (veda.exe).
	err := svc.Run("VedaEngine", &vedaService{})
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}
