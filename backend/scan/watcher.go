package scan

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
)

// StartWatcher monitors the root directory for new or modified images and
// updates the database accordingly. It runs until the process exits.
func StartWatcher(root string, gdb *gorm.DB) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watcher: %v", err)
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
				log.Printf("watcher add %s: %v", path, werr)
			}
		}
		return nil
	})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					if werr := watcher.Add(event.Name); werr != nil {
						log.Printf("watcher add %s: %v", event.Name, werr)
					}
					continue
				}
				ext := strings.ToLower(filepath.Ext(event.Name))
				switch ext {
				case ".png", ".jpg", ".jpeg", ".webp":
					go func(p string) {
						time.Sleep(500 * time.Millisecond)
						if _, err := ScanFile(gdb, root, p); err != nil {
							log.Printf("scan file %s: %v", p, err)
						}
					}(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}
