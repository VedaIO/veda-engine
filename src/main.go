//go:generate goversioninfo -64

package main

import (
	"log"

	"golang.org/x/sys/windows/svc"
)

func main() {
	// veda-anchor-engine runs exclusively as a Windows Service.
	// It is registered and started by the veda-anchor launcher (veda-anchor.exe).
	err := svc.Run("VedaAnchorEngine", &vedaAnchorService{})
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}
