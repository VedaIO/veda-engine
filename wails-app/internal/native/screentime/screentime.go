//go:build windows

package screentime

/*
#include <windows.h>
#include <stdint.h>

typedef struct {
    uint32_t pid;
    wchar_t title[256];
} ActiveWindowInfo;

int GetActiveWindowInfo(ActiveWindowInfo* info) {
    if (info == NULL) {
        return -1;
    }

    // Initialize struct
    info->pid = 0;
    info->title[0] = L'\0';

    // Get the foreground window handle
    HWND hwnd = GetForegroundWindow();
    if (hwnd == NULL) {
        return -1;
    }

    // Get the process ID of the foreground window
    DWORD pid = 0;
    GetWindowThreadProcessId(hwnd, &pid);
    if (pid == 0) {
        return -1;
    }
    info->pid = (uint32_t)pid;

    // Get the window title
    GetWindowTextW(hwnd, info->title, 256);

    return 0;
}
*/
import "C"
import "unsafe"

// ActiveWindowInfo contains information about the currently focused window
type ActiveWindowInfo struct {
	PID   uint32
	Title string
}

// GetActiveWindowInfo retrieves the PID and title of the foreground window.
// Returns nil if no foreground window is found.
func GetActiveWindowInfo() *ActiveWindowInfo {
	var cInfo C.ActiveWindowInfo
	result := C.GetActiveWindowInfo(&cInfo)
	if result != 0 {
		return nil
	}

	// Convert wide char title to Go string
	title := wcharToString(cInfo.title[:])

	return &ActiveWindowInfo{
		PID:   uint32(cInfo.pid),
		Title: title,
	}
}

// wcharToString converts a []C.wchar_t to a Go string
func wcharToString(wchars []C.wchar_t) string {
	// Find null terminator
	length := 0
	for i, c := range wchars {
		if c == 0 {
			length = i
			break
		}
	}
	if length == 0 {
		return ""
	}

	// Convert UTF-16 to UTF-8
	// Each wchar_t is 2 bytes (UTF-16LE on Windows)
	utf16 := make([]uint16, length)
	for i := 0; i < length; i++ {
		utf16[i] = uint16(wchars[i])
	}

	// Use syscall to convert UTF-16 to UTF-8
	return utf16ToString(utf16)
}

// utf16ToString converts a UTF-16 slice to a Go string
func utf16ToString(s []uint16) string {
	// Simple conversion - works for BMP characters
	runes := make([]rune, len(s))
	for i, v := range s {
		runes[i] = rune(v)
	}
	return string(runes)
}

// Stub for build system
func init() {
	// This file only compiles on Windows due to build tag
	_ = unsafe.Pointer(nil) // Silence unused import warning
}
