package logger

import (
	"io"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

var l zerolog.Logger

// Init configures the global logger. It should be called before any logging.
func Init() {
	var w io.Writer = os.Stdout
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}
	l = zerolog.New(w).With().Timestamp().Logger()
}

// Info starts a new info level log event.
func Info() *zerolog.Event {
	return l.Info()
}

// Warn starts a new warn level log event.
func Warn() *zerolog.Event {
	return l.Warn()
}

// Error starts a new error level log event.
func Error() *zerolog.Event {
	return l.Error()
}

// With returns a context for adding fields to the logger.
func With() zerolog.Context {
	return l.With()
}
