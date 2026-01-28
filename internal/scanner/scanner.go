// Package scanner provides local filesystem scanning functionality.
package scanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"godatacleaner/internal/models"
)

// Scanner scans local directories for files.
type Scanner struct {
	basePath   string
	categories []string // ["4k", "movies", "shows"]
}

// NewScanner creates a new scanner for the given base path.
func NewScanner(basePath string) *Scanner {
	return &Scanner{
		basePath:   basePath,
		categories: []string{"4k", "movies", "shows"},
	}
}

// Scan recursively scans the directory and returns files via channel.
// It uses filepath.WalkDir for efficient recursive traversal.
// Hidden files (starting with ".") are ignored.
// Context cancellation is supported for graceful shutdown.
func (s *Scanner) Scan(ctx context.Context) (<-chan models.LocalFile, <-chan error) {
	files := make(chan models.LocalFile)
	errs := make(chan error, 1)

	go func() {
		defer close(files)
		defer close(errs)

		err := filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Handle walk errors
			if err != nil {
				return err
			}

			// Get the file/directory name
			name := d.Name()

			// Skip hidden files and directories
			if isHidden(name) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip directories, we only want files
			if d.IsDir() {
				return nil
			}

			// Get file info for size
			info, err := d.Info()
			if err != nil {
				return err
			}

			// Create LocalFile and send to channel
			localFile := models.LocalFile{
				FilePath: path,
				FileName: name,
				Size:     info.Size(),
				Category: s.categorize(path),
			}

			// Send file to channel, respecting context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case files <- localFile:
			}

			return nil
		})

		if err != nil {
			// Send error to error channel (non-blocking since buffer size is 1)
			select {
			case errs <- err:
			default:
			}
		}
	}()

	return files, errs
}

// categorize determines the category of a file based on its path.
// It checks if the path contains "/4k/", "/movies/", or "/shows/".
// If none of these patterns match, it returns "unknown".
func (s *Scanner) categorize(path string) string {
	// Normalize path separators for cross-platform compatibility
	normalizedPath := filepath.ToSlash(path)

	// Check for each category in the path
	for _, category := range s.categories {
		// Check for category as a directory component (e.g., "/4k/", "/movies/", "/shows/")
		pattern := "/" + category + "/"
		if strings.Contains(normalizedPath, pattern) {
			return category
		}
	}

	return "unknown"
}

// isHidden checks if a file or directory is hidden (starts with a dot).
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
