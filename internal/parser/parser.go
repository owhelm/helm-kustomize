package parser

import (
	"bytes"
	"fmt"
	"io"

	"go.yaml.in/yaml/v4"
)

const (
	APIVersion = "helm.plugin.kustomize/v1"
	Kind       = "KustomizePluginData"
)

// KustomizePluginData represents the special resource containing kustomize files
type KustomizePluginData struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]any    `yaml:"metadata"`
	Files      map[string]string `yaml:"files"`
}

// ParseResult contains the parsed manifests separated by type
type ParseResult struct {
	KustomizePluginData *KustomizePluginData
	OtherResources      []map[string]any
}

// IsKustomizePluginDataResource checks if a resource is a KustomizePluginData resource
func IsKustomizePluginDataResource(resource map[string]any) bool {
	apiVersion, ok := resource["apiVersion"].(string)
	if !ok || apiVersion != APIVersion {
		return false
	}

	kind, ok := resource["kind"].(string)
	if !ok || kind != Kind {
		return false
	}

	return true
}

// ParseManifests parses YAML input from bytes and separates KustomizePluginData from other resources
func ParseManifests(data []byte) (*ParseResult, error) {
	result := &ParseResult{
		OtherResources: make([]map[string]any, 0),
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

		if IsKustomizePluginDataResource(doc) {
			// Parse as KustomizePluginData
			// We marshal->unmarshal because yaml.v3 doesn't provide direct conversion
			// from map[string]any to a struct. This is the idiomatic Go approach.
			var kf KustomizePluginData
			docBytes, err := yaml.Marshal(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal KustomizePluginData doc: %w", err)
			}
			if err := yaml.Unmarshal(docBytes, &kf); err != nil {
				return nil, fmt.Errorf("failed to parse KustomizePluginData resource: %w", err)
			}

			if result.KustomizePluginData != nil {
				return nil, fmt.Errorf("multiple KustomizePluginData resources found, only one is supported")
			}

			result.KustomizePluginData = &kf
		} else {
			// Keep as generic resource
			result.OtherResources = append(result.OtherResources, doc)
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
