package helpers

import "log/slog"

// Error returns an error attribute for slog logger.
func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
