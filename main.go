package main

import (
	"fmt"
	"os"

	"github.com/owhelm/helm-kustomize-plugin/internal/extractor"
	"github.com/owhelm/helm-kustomize-plugin/internal/kustomize"
	"github.com/owhelm/helm-kustomize-plugin/internal/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse input manifests
	result, err := parser.ParseManifests(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	// If no KustomizeFiles resource found, pass through the input
	if result.KustomizeFiles == nil {
		// Write all resources back to stdout
		data, err := parser.MarshalResources(result.OtherResources)
		if err != nil {
			return fmt.Errorf("failed to marshal resources: %w", err)
		}
		os.Stdout.Write(data)
		return nil
	}

	// Create temporary directory for kustomize files
	tempDir, err := extractor.NewTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer tempDir.Cleanup()

	// Extract files from KustomizeFiles resource
	if err := tempDir.ExtractFiles(result.KustomizeFiles.Files); err != nil {
		return fmt.Errorf("failed to extract files: %w", err)
	}

	// Write other resources to all.yaml
	allYamlContent, err := parser.MarshalResources(result.OtherResources)
	if err != nil {
		return fmt.Errorf("failed to marshal resources for all.yaml: %w", err)
	}

	if err := tempDir.WriteFile("all.yaml", allYamlContent); err != nil {
		return fmt.Errorf("failed to write all.yaml: %w", err)
	}

	// Check if kustomization.yaml exists and update it if needed
	kustomizationPath := "kustomization.yaml"
	kustomizationContent, err := tempDir.ReadFile(kustomizationPath)
	if err == nil {
		// kustomization.yaml exists, ensure all.yaml is in resources
		updated, changed, err := kustomize.EnsureAllYamlInKustomization(kustomizationContent)
		if err != nil {
			return fmt.Errorf("failed to update kustomization.yaml: %w", err)
		}

		if changed {
			// Write updated kustomization.yaml back
			if err := tempDir.WriteFile(kustomizationPath, updated); err != nil {
				return fmt.Errorf("failed to write updated kustomization.yaml: %w", err)
			}
		}
	}
	// If kustomization.yaml doesn't exist, that's fine - kustomize will handle it

	// Run kubectl kustomize
	output, err := kustomize.Build(tempDir.Path)
	if err != nil {
		return fmt.Errorf("failed to run kustomize: %w", err)
	}

	// Write output to stdout
	os.Stdout.Write(output)

	return nil
}
