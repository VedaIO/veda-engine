package app_filter

// Filter provides methods to determine if a process should be excluded.
// Note: ShouldTrack is now handled by Agent - see agent-migration.md
type Filter interface {
	// ShouldExclude returns true if the process is a system component that should be ignored.
	ShouldExclude(exePath string, proc any) bool
}
