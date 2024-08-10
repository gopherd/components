package loggerapi

import (
	"log/slog"
)

// go install github.com/gopherd/tools/cmd/eventer@latest
//go:generate eventer

// SetLevelEvent is an event that sets the log level.
type SetLevelEvent struct {
	Level slog.Level
}
