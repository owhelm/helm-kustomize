package main

import (
	"bytes"
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

	outputStr := output.String()
	if !strings.Contains(outputStr, "Service") {
		t.Error("Expected Service in output")
	}
	if !strings.Contains(outputStr, "Deployment") {
		t.Error("Expected Deployment in output")
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
