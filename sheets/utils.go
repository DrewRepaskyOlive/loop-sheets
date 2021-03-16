package sheets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// GetMyDocumentsPath Get path to My Documents. Verifies the path exists.
func GetMyDocumentsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	docsPath := filepath.Join(home, "Documents")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("expected path to My Documents does not exist %q: %w", docsPath, err)
	}

	return docsPath, nil
}

func GetFilesByExtension(dir, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not access path %q: %w", path, err)
		}
		if ext != filepath.Ext(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
