package gobundle

import (
	"path/filepath"
)

func ConfigFile(path string) string {
	return filepath.Join(*Setup.Directories.Config, path)
}

func DataFile(path string) string {
	return filepath.Join(*Setup.Directories.Data, path)
}