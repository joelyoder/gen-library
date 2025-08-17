package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // register webp decoder
)

func ThumbPath(sha string, width int) string {
	return filepath.ToSlash(fmt.Sprintf(".cache/thumbs/%s_%d.jpg", sha, width))
}

// EnsureThumb ensures a resized thumbnail exists for the given image.
// It lazily generates the thumbnail if missing and returns the path.
func EnsureThumb(sha string, srcPath string, width int) (string, error) {
	p := ThumbPath(sha, width)
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	img, err := imaging.Open(srcPath)
	if err != nil {
		return "", err
	}
	thumb := imaging.Resize(img, width, 0, imaging.Lanczos)
	if err := imaging.Save(thumb, p); err != nil {
		return "", err
	}
	return p, nil
}
