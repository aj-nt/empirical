package empirical

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed ephe/*
var epheFiles embed.FS

// EnsureEpheCache extracts embedded ephemeris files to ~/.cache/empirical/ephe/
// and returns the path. On subsequent calls it skips extraction if the cache
// already contains files.
func EnsureEpheCache() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	cacheDir := filepath.Join(home, ".cache", "empirical", "ephe")

	// Quick check: if cache dir exists with files, skip extraction
	if entries, _ := os.ReadDir(cacheDir); len(entries) > 0 {
		return cacheDir, nil
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return cacheDir, fs.WalkDir(epheFiles, "ephe", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(epheFiles, path)
		if err != nil {
			return err
		}

		// path is like "ephe/seas_18.se1" — extract just the filename
		targetPath := filepath.Join(cacheDir, filepath.Base(path))
		return os.WriteFile(targetPath, data, 0644)
	})
}
