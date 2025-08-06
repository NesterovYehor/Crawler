package testutils

import (
	"bytes"
	"log/slog"
)

func NewTestLogger() *bytes.Buffer {
	buf := new(bytes.Buffer)
	handler := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	looget := slog.New(handler)
	slog.SetDefault(looget)
	return buf
}
