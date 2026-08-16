package empirical

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed ephe/* ephe/ast*/*.se1
var epheFiles embed.FS

// completeMarker is written after a full extraction. Its presence (not merely
// "some files exist") signals the cache is complete. Without it, a second
// process running in parallel (e.g. `go test ./...` runs package test binaries
// concurrently) can observe a partially-extracted cache and return early,
// leaving SetEphePath pointing at an incomplete directory.
const completeMarker = ".complete"

// EnsureEpheCache extracts embedded ephemeris files to ~/.cache/empirical/ephe/
// and returns the path. Extraction is idempotent (files are overwritten), so
// concurrent callers racing to extract are safe; the completion marker only
// gates the fast-path skip.
func EnsureEpheCache() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	cacheDir := filepath.Join(home, ".cache", "empirical", "ephe")

	// Fast path: only skip if a prior extraction fully completed.
	if _, err := os.Stat(filepath.Join(cacheDir, completeMarker)); err == nil {
		return cacheDir, nil
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	err = fs.WalkDir(epheFiles, "ephe", func(path string, d fs.DirEntry, err error) error {
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

		// path is like "ephe/seas_18.se1" or "ephe/ast136/s136199s.se1"
		// Preserve subdirectory structure relative to ephe/
		relPath, err := filepath.Rel("ephe", path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(cacheDir, relPath)
		if dir := filepath.Dir(targetPath); dir != cacheDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		return os.WriteFile(targetPath, data, 0644)
	})
	if err != nil {
		return "", err
	}

	// Mark extraction complete. Written last so a concurrent reader that sees
	// the marker is guaranteed to see all files.
	if err := os.WriteFile(filepath.Join(cacheDir, completeMarker), []byte("ok"), 0644); err != nil {
		return "", err
	}
	return cacheDir, nil
}
