package logger

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

var (
	l        zerolog.Logger
	once     sync.Once
	logLevel = zerolog.InfoLevel
)

// Init configures the global logger. It should be called before any logging.
func Init() {
	once.Do(func() {
		var w io.Writer = os.Stdout
		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		}

		switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
		case "debug":
			logLevel = zerolog.DebugLevel
		case "info", "":
			logLevel = zerolog.InfoLevel
		case "warn":
			logLevel = zerolog.WarnLevel
		case "error":
			logLevel = zerolog.ErrorLevel
		default:
			logLevel = zerolog.InfoLevel
		}
		l = zerolog.New(w).Level(logLevel).With().Timestamp().Logger()
	})
}

// Level returns the configured log level.
func Level() zerolog.Level {
	return logLevel
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

// Debug starts a new debug level log event.
func Debug() *zerolog.Event {
	return l.Debug()
}

// With returns a context for adding fields to the logger.
func With() zerolog.Context {
	return l.With()
}
