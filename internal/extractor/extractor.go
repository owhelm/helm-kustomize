package extractor

import (
	"fmt"
	"os"
	"path/filepath"
)

// TempDir represents a temporary directory for kustomize files
type TempDir struct {
	Path string
	root *os.Root
}

// NewTempDir creates a new temporary directory
func NewTempDir() (*TempDir, error) {
	path, err := os.MkdirTemp("", "helm-kustomize-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempDir := &TempDir{Path: path}

	root, err := os.OpenRoot(path)
	if err != nil {
		tempDir.Cleanup()
		return nil, fmt.Errorf("failed to open root: %w", err)
	}

	tempDir.root = root
	return tempDir, nil
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
	// Create directory structure if needed
	dir := filepath.Dir(filePath)
	if err := t.root.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file content using root-constrained write
	if err := t.root.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ReadFile reads a file from the temporary directory
func (t *TempDir) ReadFile(filePath string) ([]byte, error) {
	// Read file content using root-constrained read
	content, err := t.root.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}
