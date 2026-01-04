package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestKustomizePostRenderer_Run_PassThrough(t *testing.T) {
	// Test that manifests without KustomizePluginData are passed through unchanged
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	if output.String() != input.String() {
		t.Fatalf("Run() output = %v, want %v", output, input)
	}
}

func TestKustomizePostRenderer_Run_InvalidYAML(t *testing.T) {
	input := bytes.NewBufferString(`---
invalid: yaml: structure:
  bad indentation
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestKustomizePostRenderer_Run_EmptyInput(t *testing.T) {
	input := bytes.NewBufferString("")

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	if output.Len() != 0 {
		t.Errorf("Expected empty output, got %d bytes", output.Len())
	}
}

func TestKustomizePostRenderer_Run_ReservedAllYamlFilename(t *testing.T) {
	// Test that using the reserved filename "all.yaml" returns an error
	input := bytes.NewBufferString(`---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  all.yaml: |
    some content
  kustomization.yaml: |
    resources:
      - all.yaml
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for reserved 'all.yaml' filename, got nil")
	}
	if !strings.Contains(err.Error(), "all.yaml") || !strings.Contains(err.Error(), "reserved") {
		t.Errorf("Expected error message about reserved 'all.yaml', got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_SuccessfulTransformation(t *testing.T) {
	// Test successful kustomize transformation with KustomizePluginData
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
data:
  key: value
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
    labels:
    - includeSelectors: true
      includeTemplates: true
      pairs:
        app: test-app
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	expected := `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  labels:
    app: test-app
  name: test-configmap
`

	if output.String() != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output.String())
	}
}

func TestKustomizePostRenderer_Run_KustomizationYamlUpdated(t *testing.T) {
	// Test that kustomization.yaml is updated to include all.yaml if it's missing
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
    - port: 80
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    namespace: test-namespace
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	expected := `apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: test-namespace
spec:
  ports:
  - port: 80
`

	if output.String() != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output.String())
	}
}

func TestKustomizePostRenderer_Run_WithPatches(t *testing.T) {
	// Test kustomize transformation with patches
	input := bytes.NewBufferString(`---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
    patches:
      - patch: |-
          - op: replace
            path: /spec/replicas
            value: 3
        target:
          kind: Deployment
          name: test-deployment
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
`

	if output.String() != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output.String())
	}
}

func TestKustomizePostRenderer_Run_MultipleResources(t *testing.T) {
	// Test with multiple resources and kustomize transformations
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  config: value
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
    commonAnnotations:
      managed-by: helm-kustomize-plugin
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	expected := `apiVersion: v1
data:
  config: value
kind: ConfigMap
metadata:
  annotations:
    managed-by: helm-kustomize-plugin
  name: my-config
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    managed-by: helm-kustomize-plugin
  name: my-service
spec:
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    managed-by: helm-kustomize-plugin
  name: my-deployment
spec:
  replicas: 2
  template:
    metadata:
      annotations:
        managed-by: helm-kustomize-plugin
`

	if output.String() != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output.String())
	}
}

func TestKustomizePostRenderer_Run_ExtractFilesInvalidPath(t *testing.T) {
	// Test that directory traversal in file paths is rejected
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  ../../../etc/passwd: |
    malicious content
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for directory traversal attempt, got nil")
	}
	if !strings.Contains(err.Error(), "failed to extract files") {
		t.Errorf("Expected error message about failed to extract files, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_InvalidKustomizationResources(t *testing.T) {
	// Test that invalid kustomization.yaml (resources not array) returns error
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    resources: "not an array"
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for invalid kustomization.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update kustomization.yaml") {
		t.Errorf("Expected error message about failed to update kustomization.yaml, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_InvalidKustomizationYaml(t *testing.T) {
	// Test that invalid kustomization.yaml causes kubectl kustomize to fail
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
      - nonexistent-file.yaml
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for invalid kustomization (missing resource file), got nil")
	}
	if !strings.Contains(err.Error(), "failed to run kustomize") {
		t.Errorf("Expected error message about failed to run kustomize, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_TempDirCreationFailure(t *testing.T) {
	// Test that NewTempDir() failure is handled correctly
	// We'll manipulate TMPDIR to point to a read-only directory
	originalTmpDir := os.Getenv("TMPDIR")
	defer func() {
		if originalTmpDir != "" {
			if err := os.Setenv("TMPDIR", originalTmpDir); err != nil {
				t.Logf("Failed to restore TMPDIR: %v", err)
			}
		} else {
			if err := os.Unsetenv("TMPDIR"); err != nil {
				t.Logf("Failed to unset TMPDIR: %v", err)
			}
		}
	}()

	// Create a temporary read-only directory
	readOnlyDir, err := os.MkdirTemp("", "readonly-*")
	if err != nil {
		t.Skipf("Cannot create test directory: %v", err)
	}
	defer func() {
		// Restore permissions before cleanup
		if err := os.Chmod(readOnlyDir, 0755); err != nil {
			t.Logf("Failed to restore permissions: %v", err)
		}
		if err := os.RemoveAll(readOnlyDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Make it read-only
	if err := os.Chmod(readOnlyDir, 0444); err != nil {
		t.Skipf("Cannot chmod directory: %v", err)
	}

	// Set TMPDIR to the read-only directory
	if err := os.Setenv("TMPDIR", readOnlyDir); err != nil {
		t.Fatalf("Failed to set TMPDIR: %v", err)
	}

	// Prepare input with KustomizePluginData
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
`)

	renderer := &KustomizePostRenderer{}
	_, err = renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error when temp directory creation fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create temp directory") {
		t.Errorf("Expected error message about failed to create temp directory, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_NoFilesSection(t *testing.T) {
	// Test that KustomizePluginData without files section returns an error
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for KustomizePluginData without files section, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse input") && !strings.Contains(err.Error(), "'files' field must be a map") {
		t.Errorf("Expected error message about files field, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_NoKustomizationYaml(t *testing.T) {
	// Test that files section without kustomization.yaml returns an error
	input := bytes.NewBufferString(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  patch.yaml: |
    apiVersion: v1
    kind: Service
    metadata:
      name: test-service
      labels:
        patched: "true"
`)

	renderer := &KustomizePostRenderer{}
	_, err := renderer.Run(input)
	if err == nil {
		t.Fatal("Expected error for files without kustomization.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "failed to run kustomize") && !strings.Contains(err.Error(), "unable to find") {
		t.Errorf("Expected error message about missing kustomization.yaml, got: %v", err)
	}
}

func TestKustomizePostRenderer_Run_NestedPaths(t *testing.T) {
	// Test that files with nested directory paths are handled correctly
	input := bytes.NewBufferString(`---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test
        image: nginx:latest
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - all.yaml
    patches:
      - path: patches/replica-patch.yaml
  patches/replica-patch.yaml: |
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: test-deployment
    spec:
      replicas: 5
`)

	renderer := &KustomizePostRenderer{}
	output, err := renderer.Run(input)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 5
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - image: nginx:latest
        name: test
`

	if output.String() != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output.String())
	}
}
