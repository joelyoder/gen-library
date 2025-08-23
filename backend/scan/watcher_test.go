package scan

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gen-library/backend/db"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.ApplyMigrations(gdb))
	return gdb
}

func createPNG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, png.Encode(f, img))
}

func TestWatcherStartStop(t *testing.T) {
	gdb := newTestDB(t)
	root := t.TempDir()

	done := make(chan struct{})
	go func() {
		StartWatcher(root, gdb)
		close(done)
	}()
	t.Cleanup(func() {
		StopWatcher()
		<-done
	})

	require.Eventually(t, IsWatcherRunning, time.Second, 10*time.Millisecond)

	StopWatcher()
	<-done
	require.False(t, IsWatcherRunning())
}

func TestWatcherScansNewFile(t *testing.T) {
	gdb := newTestDB(t)
	root := t.TempDir()

	done := make(chan struct{})
	go func() {
		StartWatcher(root, gdb)
		close(done)
	}()
	t.Cleanup(func() {
		StopWatcher()
		<-done
	})

	require.Eventually(t, IsWatcherRunning, time.Second, 10*time.Millisecond)

	imgPath := filepath.Join(root, "test.png")
	createPNG(t, imgPath)

	require.Eventually(t, func() bool {
		var count int64
		gdb.Model(&db.Image{}).Count(&count)
		return count == 1
	}, 5*time.Second, 100*time.Millisecond)

	StopWatcher()
	<-done
	require.False(t, IsWatcherRunning())
}
