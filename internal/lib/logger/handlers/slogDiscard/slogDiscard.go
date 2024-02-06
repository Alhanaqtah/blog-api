package slogDiscard

import (
	"context"
	"log/slog"
)

type DiscardLogger struct{}

func NewDiscardLogger() *slog.Logger {
	return slog.New(&DiscardLogger{})
}

func (d *DiscardLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (d *DiscardLogger) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (d *DiscardLogger) WithAttrs(_ []slog.Attr) slog.Handler {
	return nil
}

func (d *DiscardLogger) WithGroup(_ string) slog.Handler {
	return nil
}
