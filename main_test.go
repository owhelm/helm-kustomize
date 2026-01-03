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
