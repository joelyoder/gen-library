package scan

import (
	"io/fs"
	"log"
	"path/filepath"

	"gorm.io/gorm"
)

func ScanFolder(gdb *gorm.DB, root string) error {
	log.Printf("scan stub: would scan %s for images (*.png,*.jpg,*.jpeg,*.webp)", root)
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { return nil })
}
