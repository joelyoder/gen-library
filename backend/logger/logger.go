package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	logFile  *os.File
)

// Init configures the global logger. It should be called before any logging.
func Init() {
	once.Do(func() {
		var w io.Writer = os.Stdout
		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		}

		if name := os.Getenv("LOG_FILE"); name != "" {
			name = filepath.Base(name)
			if err := os.MkdirAll("logs", 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "logger: %v\n", err)
			} else {
				path := filepath.Join("logs", name)
				if _, err := os.Stat(path); err == nil {
					_ = os.Rename(path, path+".1")
				}
				f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "logger: %v\n", err)
				} else {
					logFile = f
					w = io.MultiWriter(w, f)
				}
			}
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

// Close releases any resources held by the logger.
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
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
