package hlsClipper

import (
	"fmt"
	"os"
	"path/filepath"
)

// Clean Temp Directory
func cleanTempDir() error {
	err := filepath.Walk(tmpPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == tmpPath {
			return nil
		}
		if info.IsDir() {
			return os.RemoveAll(path)
		}
		return os.Remove(path)
	})

	if err != nil {
		return err
	}

	return nil
}

// Check Directories
func checkDirs() error {
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		if err := os.Mkdir(tmpPath, 0755); err != nil {
			return fmt.Errorf("| error | os.Mkdir | temp dir could not be created %v", err)
		}
	}

	if _, err := os.Stat(clipsPath); os.IsNotExist(err) {
		if err := os.Mkdir(clipsPath, 0755); err != nil {
			return fmt.Errorf("| error | os.Mkdir | clips dir could not be created %v", err)
		}
	}
	return nil
}

func checkFileExist(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("| error | os.IsNotExist | %v", err)
	}

	return fmt.Errorf("| error | clip file already exist!")
}
