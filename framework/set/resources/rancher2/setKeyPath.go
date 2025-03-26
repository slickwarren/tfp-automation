package rancher2

import (
	"os"
	"path/filepath"
)

// SetKeyPath is a function that will set the path to the key file.
func SetKeyPath(keyPath string) string {
	userDir := os.Getenv("GOROOT")

	keyPath = filepath.Join(userDir, keyPath)

	return keyPath
}
