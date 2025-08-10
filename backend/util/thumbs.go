package util

import (
	"fmt"
	"path/filepath"
)

func ThumbPath(sha string, width int) string {
	return filepath.ToSlash(fmt.Sprintf(".cache/thumbs/%s_%d.jpg", sha, width))
}

// EnsureThumb is a placeholder for future lazy generation.
// For the MVP we only compute the path; generation will be wired later.
func EnsureThumb(sha string, srcPath string, width int) (string, error) {
	return ThumbPath(sha, width), nil
}
