package scan

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func ScanFolder(gdb *gorm.DB, root string) (int, int, error) {
	if root == "" {
		return 0, 0, nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			log.Printf("ScanFolder: folder %s does not exist", root)
			return 0, 0, nil
		}
		return 0, 0, err
	}
	log.Printf("scan stub: would scan %s for images (*.png,*.jpg,*.jpeg,*.webp)", root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { return nil })
	return 0, 0, err
}
