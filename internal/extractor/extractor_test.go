package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTempDir(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v, want nil", err)
	}
	defer tempDir.Cleanup()

	if tempDir.Path == "" {
		t.Error("Expected non-empty path")
	}

	// Verify directory exists
	if _, err := os.Stat(tempDir.Path); os.IsNotExist(err) {
		t.Errorf("Directory does not exist: %s", tempDir.Path)
	}
}

func TestTempDir_Cleanup(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}

	path := tempDir.Path

	// Verify directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Directory should exist: %s", path)
	}

	// Cleanup
	tempDir.Cleanup()

	// Verify directory no longer exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Directory should not exist after cleanup: %s", path)
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple filename",
			path:    "file.yaml",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "patches/deployment.yaml",
			wantErr: false,
		},
		{
			name:    "deeply nested path",
			path:    "overlays/production/patches/deployment.yaml",
			wantErr: false,
		},
		{
			name:    "parent directory traversal",
			path:    "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "traversal in middle",
			path:    "foo/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "current directory reference",
			path:    "./file.yaml",
			wantErr: false,
		},
		{
			name:    "multiple slashes",
			path:    "foo//bar.yaml",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTempDir_ExtractFiles(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	files := map[string]string{
		"kustomization.yaml":       "resources:\n- all.yaml\n",
		"patches/deployment.yaml":  "apiVersion: apps/v1\nkind: Deployment\n",
		"overlays/prod/patch.yaml": "spec:\n  replicas: 3\n",
	}

	err = tempDir.ExtractFiles(files)
	if err != nil {
		t.Fatalf("ExtractFiles() error = %v, want nil", err)
	}

	// Verify all files were created
	for filePath, expectedContent := range files {
		fullPath := filepath.Join(tempDir.Path, filePath)

		// Check file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("File should exist: %s", filePath)
			continue
		}

		// Check content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filePath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("File %s content = %q, want %q", filePath, string(content), expectedContent)
		}
	}
}

func TestTempDir_ExtractFiles_InvalidPath(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	files := map[string]string{
		"../../../etc/passwd": "malicious content",
	}

	err = tempDir.ExtractFiles(files)
	if err == nil {
		t.Fatal("ExtractFiles() should return error for directory traversal attempt")
	}
}

func TestTempDir_WriteFile(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	content := []byte("test content")
	err = tempDir.WriteFile("test.yaml", content)
	if err != nil {
		t.Fatalf("WriteFile() error = %v, want nil", err)
	}

	// Verify file was created
	fullPath := filepath.Join(tempDir.Path, "test.yaml")
	readContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("File content = %q, want %q", string(readContent), string(content))
	}
}

func TestTempDir_WriteFile_NestedPath(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	content := []byte("nested content")
	err = tempDir.WriteFile("subdir/nested/test.yaml", content)
	if err != nil {
		t.Fatalf("WriteFile() error = %v, want nil", err)
	}

	// Verify file was created
	fullPath := filepath.Join(tempDir.Path, "subdir/nested/test.yaml")
	readContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("File content = %q, want %q", string(readContent), string(content))
	}
}

func TestTempDir_WriteFile_InvalidPath(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	err = tempDir.WriteFile("../../../etc/passwd", []byte("malicious"))
	if err == nil {
		t.Fatal("WriteFile() should return error for directory traversal attempt")
	}
}

func TestTempDir_ReadFile(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	// Write a file first
	expectedContent := []byte("test content")
	err = tempDir.WriteFile("test.yaml", expectedContent)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Read it back
	content, err := tempDir.ReadFile("test.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v, want nil", err)
	}

	if string(content) != string(expectedContent) {
		t.Errorf("ReadFile() = %q, want %q", string(content), string(expectedContent))
	}
}

func TestTempDir_ReadFile_NotExists(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	_, err = tempDir.ReadFile("nonexistent.yaml")
	if err == nil {
		t.Fatal("ReadFile() should return error for nonexistent file")
	}
}

func TestTempDir_ReadFile_InvalidPath(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	_, err = tempDir.ReadFile("../../../etc/passwd")
	if err == nil {
		t.Fatal("ReadFile() should return error for directory traversal attempt")
	}
}
