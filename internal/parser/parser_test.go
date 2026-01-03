package parser

import (
	"strings"
	"testing"
)

func TestIsKustomizeFilesResource(t *testing.T) {
	tests := []struct {
		name     string
		resource map[string]any
		want     bool
	}{
		{
			name: "valid KustomizeFiles resource",
			resource: map[string]any{
				"apiVersion": "helm.kustomize.plugin/v1alpha1",
				"kind":       "KustomizeFiles",
			},
			want: true,
		},
		{
			name: "wrong apiVersion",
			resource: map[string]any{
				"apiVersion": "v1",
				"kind":       "KustomizeFiles",
			},
			want: false,
		},
		{
			name: "wrong kind",
			resource: map[string]any{
				"apiVersion": "helm.kustomize.plugin/v1alpha1",
				"kind":       "ConfigMap",
			},
			want: false,
		},
		{
			name: "missing apiVersion",
			resource: map[string]any{
				"kind": "KustomizeFiles",
			},
			want: false,
		},
		{
			name: "missing kind",
			resource: map[string]any{
				"apiVersion": "helm.kustomize.plugin/v1alpha1",
			},
			want: false,
		},
		{
			name:     "empty resource",
			resource: map[string]any{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKustomizeFilesResource(tt.resource)
			if got != tt.want {
				t.Errorf("IsKustomizeFilesResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseManifests_EmptyInput(t *testing.T) {
	input := strings.NewReader("")
	result, err := ParseManifests(input)

	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizeFiles != nil {
		t.Errorf("Expected no KustomizeFiles, got %v", result.KustomizeFiles)
	}

	if len(result.OtherResources) != 0 {
		t.Errorf("Expected no OtherResources, got %d", len(result.OtherResources))
	}
}

func TestParseManifests_NoKustomizeFiles(t *testing.T) {
	input := strings.NewReader(`---
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

	result, err := ParseManifests(input)
	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizeFiles != nil {
		t.Errorf("Expected no KustomizeFiles, got %v", result.KustomizeFiles)
	}

	if len(result.OtherResources) != 2 {
		t.Errorf("Expected 2 OtherResources, got %d", len(result.OtherResources))
	}

	if len(result.RawOthers) != 2 {
		t.Errorf("Expected 2 RawOthers, got %d", len(result.RawOthers))
	}
}

func TestParseManifests_WithKustomizeFiles(t *testing.T) {
	input := strings.NewReader(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.kustomize.plugin/v1alpha1
kind: KustomizeFiles
metadata:
  name: kustomize-files
files:
  kustomization.yaml: |
    resources:
    - all.yaml
  patch.yaml: |
    apiVersion: apps/v1
    kind: Deployment
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`)

	result, err := ParseManifests(input)
	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizeFiles == nil {
		t.Fatal("Expected KustomizeFiles, got nil")
	}

	if result.KustomizeFiles.APIVersion != "helm.kustomize.plugin/v1alpha1" {
		t.Errorf("Expected apiVersion helm.kustomize.plugin/v1alpha1, got %s", result.KustomizeFiles.APIVersion)
	}

	if result.KustomizeFiles.Kind != "KustomizeFiles" {
		t.Errorf("Expected kind KustomizeFiles, got %s", result.KustomizeFiles.Kind)
	}

	if len(result.KustomizeFiles.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.KustomizeFiles.Files))
	}

	if _, ok := result.KustomizeFiles.Files["kustomization.yaml"]; !ok {
		t.Error("Expected kustomization.yaml file")
	}

	if _, ok := result.KustomizeFiles.Files["patch.yaml"]; !ok {
		t.Error("Expected patch.yaml file")
	}

	// Should have 2 other resources (Service and Deployment)
	if len(result.OtherResources) != 2 {
		t.Errorf("Expected 2 OtherResources, got %d", len(result.OtherResources))
	}

	if len(result.RawOthers) != 2 {
		t.Errorf("Expected 2 RawOthers, got %d", len(result.RawOthers))
	}
}

func TestParseManifests_MultipleKustomizeFiles(t *testing.T) {
	input := strings.NewReader(`---
apiVersion: helm.kustomize.plugin/v1alpha1
kind: KustomizeFiles
metadata:
  name: first
files:
  file1.yaml: content1
---
apiVersion: helm.kustomize.plugin/v1alpha1
kind: KustomizeFiles
metadata:
  name: second
files:
  file2.yaml: content2
`)

	_, err := ParseManifests(input)
	if err == nil {
		t.Fatal("Expected error for multiple KustomizeFiles, got nil")
	}

	if !strings.Contains(err.Error(), "multiple KustomizeFiles") {
		t.Errorf("Expected error about multiple KustomizeFiles, got: %v", err)
	}
}

func TestParseManifests_InvalidYAML(t *testing.T) {
	input := strings.NewReader(`---
this is not: valid: yaml: structure
  bad indentation
`)

	_, err := ParseManifests(input)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestMarshalResources(t *testing.T) {
	resources := []map[string]any{
		{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name": "test-service",
			},
		},
		{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "test-deployment",
			},
		},
	}

	data, err := MarshalResources(resources)
	if err != nil {
		t.Fatalf("MarshalResources() error = %v, want nil", err)
	}

	str := string(data)

	// Should contain document separator
	if !strings.Contains(str, "---") {
		t.Error("Expected document separator in output")
	}

	// Should contain both resources
	if !strings.Contains(str, "Service") {
		t.Error("Expected Service in output")
	}

	if !strings.Contains(str, "Deployment") {
		t.Error("Expected Deployment in output")
	}
}

func TestMarshalResources_Empty(t *testing.T) {
	var resources []map[string]any

	data, err := MarshalResources(resources)
	if err != nil {
		t.Fatalf("MarshalResources() error = %v, want nil", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty output, got %d bytes", len(data))
	}
}
