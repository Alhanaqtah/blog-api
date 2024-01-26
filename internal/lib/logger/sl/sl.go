package sl

import "log/slog"

// Error wraps errors for slog
func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
