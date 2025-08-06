package utils

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewProdLogger(path string, maxSizeMB, maxBackups, maxAgeDays int) func() {
	rotator := &lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    100, // MB
		MaxBackups: 7,
		MaxAge:     30, // days
	}

	asyncH := NewAsyncHandler(
		rotator,
		4*1024,
		1000,
		5*time.Second,
	)

	switcher := &LevelSwitcher{
		direct:   slog.NewJSONHandler(rotator, nil),
		buffered: asyncH,
	}
	logger := slog.New(switcher)
	slog.SetDefault(logger)
	return asyncH.Close
}

type LevelSwitcher struct {
	direct   slog.Handler
	buffered slog.Handler
}

func (h LevelSwitcher) Enabled(ctx context.Context, r slog.Level) bool {
	// you could consult r.Level here if you wanted to disable e.g. DEBUG entirely
	return true
}

func (h LevelSwitcher) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		// ERROR+ goes straight to disk
		return h.direct.Handle(ctx, r)
	}
	// INFO/WARN/etc go into the buffer
	return h.buffered.Handle(ctx, r)
}

func (h LevelSwitcher) WithAttrs(attrs []slog.Attr) slog.Handler {
	return LevelSwitcher{
		direct:   h.direct.WithAttrs(attrs),
		buffered: h.buffered.WithAttrs(attrs),
	}
}

func (h LevelSwitcher) WithGroup(name string) slog.Handler {
	return LevelSwitcher{
		direct:   h.direct.WithGroup(name),
		buffered: h.buffered.WithGroup(name),
	}
}

type AsyncHandler struct {
	ch       chan []byte    // raw serialized record bytes
	wg       sync.WaitGroup // to wait for writer goroutine
	bufw     *bufio.Writer
	interval time.Duration
	done     chan struct{}
}

func NewAsyncHandler(
	writer io.Writer, // underlying lumberjack.Logger
	bufSize int, // bufio buffer size
	chanSize int, // length of the channel
	flushInterval time.Duration, // how often to force a flush
) *AsyncHandler {
	bufw := bufio.NewWriterSize(writer, bufSize)
	h := &AsyncHandler{
		ch:       make(chan []byte, chanSize),
		bufw:     bufw,
		interval: flushInterval,
		done:     make(chan struct{}),
	}

	// Start the writer goroutine
	h.wg.Add(1)
	go h.writerLoop()

	return h
}

func (h *AsyncHandler) writerLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-h.ch:
			if !ok {
				// Attempt to write the last read data
				if len(data) > 0 {
					if _, err := h.bufw.Write(data); err != nil {
						slog.Error("failed to write to buffer after channel close", "error", err)
					}
				}

				// Drain remaining data from the channel
				for d := range h.ch {
					if _, err := h.bufw.Write(d); err != nil {
						slog.Error("failed to write remaining data to buffer", "error", err)
					}
				}

				if err := h.bufw.Flush(); err != nil {
					slog.Error("failed to flush buffer on shutdown", "error", err)
				}
				return
			}

			if _, err := h.bufw.Write(data); err != nil {
				slog.Error("failed to write to buffer", "error", err)
			}

		case <-ticker.C:
			if err := h.bufw.Flush(); err != nil {
				slog.Error("failed to flush buffer", "error", err)
			}
		}
	}
}

func (h *AsyncHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer
	jsonH := slog.NewJSONHandler(&buf, nil)
	if err := jsonH.Handle(ctx, r); err != nil {
		return err
	}
	buf.WriteByte('\n')
	data := buf.Bytes()
	select {
	case h.ch <- data:
		return nil
	default:
		<-h.ch
		h.ch <- data
		return nil
	}
}

func (h *AsyncHandler) Enabled(ctx context.Context, r slog.Level) bool {
	return true
}

func (h *AsyncHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *AsyncHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *AsyncHandler) Close() {
	close(h.ch)
	h.wg.Wait()
}
