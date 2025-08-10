package scan

import (
        "io/fs"
        "log"
        "os"
        "path/filepath"

        "gorm.io/gorm"
)

// ScanFolder walks the given root folder and imports images.
// Returns counts of added and updated images.
func ScanFolder(gdb *gorm.DB, root string) (int, int, error) {
        if root == "" {
                log.Printf("scan skipped: no folder configured")
                return 0, 0, nil
        }
        if info, err := os.Stat(root); err != nil || !info.IsDir() {
                if err != nil {
                        log.Printf("scan skipped: %v", err)
                } else {
                        log.Printf("scan skipped: %s is not a directory", root)
                }
                return 0, 0, nil
        }
        log.Printf("scan stub: would scan %s for images (*.png,*.jpg,*.jpeg,*.webp)", root)
        if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { return nil }); err != nil {
                return 0, 0, err
        }
        return 0, 0, nil
}
