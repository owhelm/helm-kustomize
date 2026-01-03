package parser

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// KustomizeFiles represents the special resource containing kustomize files
type KustomizeFiles struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]any    `yaml:"metadata"`
	Files      map[string]string `yaml:"files"`
}

// ParseResult contains the parsed manifests separated by type
type ParseResult struct {
	KustomizeFiles *KustomizeFiles
	OtherResources []map[string]any
	RawOthers      [][]byte // Raw YAML bytes for other resources
}

// IsKustomizeFilesResource checks if a resource is a KustomizeFiles resource
func IsKustomizeFilesResource(resource map[string]any) bool {
	apiVersion, ok := resource["apiVersion"].(string)
	if !ok || apiVersion != "helm.kustomize.plugin/v1alpha1" {
		return false
	}

	kind, ok := resource["kind"].(string)
	if !ok || kind != "KustomizeFiles" {
		return false
	}

	return true
}

// ParseManifests reads YAML input and separates KustomizeFiles from other resources
func ParseManifests(input io.Reader) (*ParseResult, error) {
	result := &ParseResult{
		OtherResources: make([]map[string]any, 0),
		RawOthers:      make([][]byte, 0),
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Split by YAML document separator
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var doc map[string]any
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML document: %w", err)
		}

		// Skip empty documents
		if doc == nil || len(doc) == 0 {
			continue
		}

		if IsKustomizeFilesResource(doc) {
			// Parse as KustomizeFiles
			// We marshal->unmarshal because yaml.v3 doesn't provide direct conversion
			// from map[string]any to a struct. This is the idiomatic Go approach.
			var kf KustomizeFiles
			docBytes, err := yaml.Marshal(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal KustomizeFiles doc: %w", err)
			}
			if err := yaml.Unmarshal(docBytes, &kf); err != nil {
				return nil, fmt.Errorf("failed to parse KustomizeFiles resource: %w", err)
			}

			if result.KustomizeFiles != nil {
				return nil, fmt.Errorf("multiple KustomizeFiles resources found, only one is supported")
			}

			result.KustomizeFiles = &kf
		} else {
			// Keep as generic resource
			result.OtherResources = append(result.OtherResources, doc)

			// Also keep raw bytes for exact reproduction
			docBytes, err := yaml.Marshal(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal resource: %w", err)
			}
			result.RawOthers = append(result.RawOthers, docBytes)
		}
	}

	return result, nil
}

// MarshalResources converts resources back to YAML format
func MarshalResources(resources []map[string]any) ([]byte, error) {
	var buf bytes.Buffer

	for i, resource := range resources {
		if i > 0 {
			buf.WriteString("---\n")
		}

		data, err := yaml.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resource: %w", err)
		}

		buf.Write(data)
	}

	return buf.Bytes(), nil
}
