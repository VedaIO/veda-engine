package config

const (
	// AppName is used for directories (e.g. in AppData)
	AppName = "VedaAnchor"

	// ServiceName is the Windows Service name
	ServiceName = "VedaAnchorEngine"

	// PipeName is the named pipe address for IPC
	PipeName = `\\.\pipe\veda-anchor`
)
