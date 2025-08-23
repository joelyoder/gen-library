package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
)

var (
	watchMu sync.Mutex
	cancel  context.CancelFunc
)

// StartWatcher monitors the root directory for new or modified images and
// updates the database accordingly. It runs until StopWatcher is called.
func StartWatcher(root string, gdb *gorm.DB) {
	watchMu.Lock()
	if cancel != nil {
		watchMu.Unlock()
		return // already running
	}
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	watchMu.Unlock()

	runWatcher(ctx, root, gdb)

	watchMu.Lock()
	cancel = nil
	watchMu.Unlock()
}

// StopWatcher stops the background watcher if it's running.
func StopWatcher() {
	watchMu.Lock()
	defer watchMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// IsWatcherRunning returns true if the watcher is currently active.
func IsWatcherRunning() bool {
	watchMu.Lock()
	defer watchMu.Unlock()
	return cancel != nil
}

func runWatcher(ctx context.Context, root string, gdb *gorm.DB) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.With("component", "scan", "event", "watcher").Error("watcher", "err", err)
		return
	}
	defer watcher.Close()

	// Recursively register existing directories
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if werr := watcher.Add(path); werr != nil {
				logger.With("component", "scan", "event", "watcher_add", "path", path).Warn("watcher_add", "err", werr)
			}
		}
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					if werr := watcher.Add(event.Name); werr != nil {
						logger.With("component", "scan", "event", "watcher_add", "path", event.Name).Warn("watcher_add", "err", werr)
					}
					continue
				}
				ext := strings.ToLower(filepath.Ext(event.Name))
				switch ext {
				case ".png", ".jpg", ".jpeg", ".webp":
					go func(p string) {
						time.Sleep(500 * time.Millisecond)
						if _, err := ScanFile(gdb, root, p); err != nil {
							logger.With("component", "scan", "path", p).Warn("scan file", "err", err)
						}
					}(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.With("component", "scan", "event", "watcher").Error("watcher", "err", err)
		}
	}
}
