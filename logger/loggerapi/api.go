package loggerapi

import "log/slog"

type Component interface {
	SetLogLevel(slog.Level)
	GetLogLevel() slog.Level
}
