package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"helm.sh/helm/v4/pkg/postrenderer"

	"github.com/owhelm/helm-kustomize-plugin/internal/extractor"
	"github.com/owhelm/helm-kustomize-plugin/internal/kustomize"
	"github.com/owhelm/helm-kustomize-plugin/internal/parser"
)

// KustomizePostRenderer implements the Helm PostRenderer interface.
// It processes manifests through kustomize transformations.
type KustomizePostRenderer struct{}

// Compile-time interface compliance check
var _ postrenderer.PostRenderer = (*KustomizePostRenderer)(nil)

func main() {
	// Create the post-renderer
	renderer := &KustomizePostRenderer{}

	// Read input from stdin into a buffer
	input := &bytes.Buffer{}
	if _, err := io.Copy(input, os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read input: %v\n", err)
		os.Exit(1)
	}

	// Process manifests using the PostRenderer interface
	output, err := renderer.Run(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write output to stdout
	if _, err := io.Copy(os.Stdout, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
		os.Exit(1)
	}
}

// Run implements the Helm PostRenderer interface.
// It processes rendered manifests through kustomize transformations.
func (k *KustomizePostRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	// Parse input manifests
	result, err := parser.ParseManifests(renderedManifests.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	// If no KustomizePluginData resource found, pass through the input
	if result.KustomizePluginData == nil {
		data, err := parser.MarshalResources(result.OtherResources)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resources: %w", err)
		}
		return bytes.NewBuffer(data), nil
	}

	// Create temporary directory for kustomize files
	tempDir, err := extractor.NewTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer tempDir.Cleanup()

	// Extract files from KustomizePluginData resource
	if err := tempDir.ExtractFiles(result.KustomizePluginData.Files); err != nil {
		return nil, fmt.Errorf("failed to extract files: %w", err)
	}

	// Write other resources to all.yaml
	allYamlContent, err := parser.MarshalResources(result.OtherResources)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources for all.yaml: %w", err)
	}

	if err := tempDir.WriteFile("all.yaml", allYamlContent); err != nil {
		return nil, fmt.Errorf("failed to write all.yaml: %w", err)
	}

	// Check if kustomization.yaml exists and update it if needed
	kustomizationPath := "kustomization.yaml"
	kustomizationContent, err := tempDir.ReadFile(kustomizationPath)
	if err == nil {
		// kustomization.yaml exists, ensure all.yaml is in resources
		updated, changed, err := kustomize.EnsureAllYamlInKustomization(kustomizationContent)
		if err != nil {
			return nil, fmt.Errorf("failed to update kustomization.yaml: %w", err)
		}

		if changed {
			// Write updated kustomization.yaml back
			if err := tempDir.WriteFile(kustomizationPath, updated); err != nil {
				return nil, fmt.Errorf("failed to write updated kustomization.yaml: %w", err)
			}
		}
	}
	// If kustomization.yaml doesn't exist, that's fine - kustomize will handle it

	// Run kubectl kustomize
	output, err := kustomize.Build(tempDir.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to run kustomize: %w", err)
	}

	return bytes.NewBuffer(output), nil
}
