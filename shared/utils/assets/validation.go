package assets

import (
	"path/filepath"
	"strings"
)

func ValidateProfileImagePath(path string) error {
	if path == "" {
		return nil
	}

	// 1. Length check
	if len(path) > 500 {
		return ErrInvalidProfileImagePath
	}

	// 2. Prevent Path Traversal attacks (e.g., "../../secret.txt")
	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
		return ErrInvalidProfileImagePath
	}

	// 3. Extension check
	ext := strings.ToLower(filepath.Ext(path))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	if !allowedExtensions[ext] {
		return ErrInvalidProfileImageType
	}

	return nil
}
