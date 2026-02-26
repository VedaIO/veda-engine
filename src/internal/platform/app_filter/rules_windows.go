//go:build windows

package app_filter

import (
	"strings"
	"veda-anchor-engine/src/internal/platform/executable"
	"veda-anchor-engine/src/internal/platform/proc_sensing"
	"veda-anchor-engine/src/internal/platform/process_integrity"
)

// ShouldExclude returns true if the process is a Windows system component, conhost.exe, or Veda Anchor itself.
func ShouldExclude(exePath string, proc *proc_sensing.ProcessInfo) bool {
	exePathLower := strings.ToLower(exePath)

	// Never track Veda Anchor itself
	if strings.Contains(exePathLower, "veda.exe") {
		return true
	}

	// Skip conhost.exe
	if strings.HasSuffix(exePathLower, "conhost.exe") {
		return true
	}

	// Skip if in System32/SysWOW64 (Windows system processes)
	if strings.Contains(exePathLower, "\\windows\\system32\\") ||
		strings.Contains(exePathLower, "\\windows\\syswow64\\") {
		return true
	}

	// Skip processes with "Microsoft® Windows® Operating System" product name
	productName, err := executable.GetProductName(exePath)
	if err == nil && strings.Contains(productName, "Microsoft® Windows® Operating System") {
		return true
	}

	// Skip system integrity level processes (system services)
	if proc != nil {
		il := process_integrity.GetProcessLevel(uint32(proc.PID))
		if il >= process_integrity.SystemRID {
			return true
		}
	}

	return false
}
