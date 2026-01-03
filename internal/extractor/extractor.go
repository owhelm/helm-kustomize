package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TempDir represents a temporary directory for kustomize files
type TempDir struct {
	Path string
}

// NewTempDir creates a new temporary directory
func NewTempDir() (*TempDir, error) {
	path, err := os.MkdirTemp("", "helm-kustomize-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &TempDir{Path: path}, nil
}

// Cleanup removes the temporary directory and all its contents.
// If cleanup fails, it prints a warning to stderr but does not return an error,
// as the OS should eventually clean up temporary files.
func (t *TempDir) Cleanup() {
	if t.Path == "" {
		return
	}

	if err := os.RemoveAll(t.Path); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup temp directory %s: %v\n", t.Path, err)
	}
}

// validatePath checks if a file path is safe (prevents directory traversal)
func validatePath(filePath string) error {
	// Clean the path to resolve any . or .. components
	cleaned := filepath.Clean(filePath)

	// Check for absolute paths
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute paths are not allowed: %s", filePath)
	}

	// Check for directory traversal attempts
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return fmt.Errorf("directory traversal not allowed: %s", filePath)
	}

	return nil
}

// ExtractFiles writes files from the files map to the temporary directory
func (t *TempDir) ExtractFiles(files map[string]string) error {
	for filePath, content := range files {
		if err := t.WriteFile(filePath, []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// WriteFile writes content to a file in the temporary directory
func (t *TempDir) WriteFile(filePath string, content []byte) error {
	// Validate the path
	if err := validatePath(filePath); err != nil {
		return err
	}

	// Create full path within temp directory
	fullPath := filepath.Join(t.Path, filePath)

	// Create directory structure if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file content
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ReadFile reads a file from the temporary directory
func (t *TempDir) ReadFile(filePath string) ([]byte, error) {
	// Validate the path
	if err := validatePath(filePath); err != nil {
		return nil, err
	}

	// Create full path within temp directory
	fullPath := filepath.Join(t.Path, filePath)

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}
