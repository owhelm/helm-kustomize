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
	Files      map[string]string `yaml:"files"`
}

// ParseResult contains the parsed manifests separated by type
type ParseResult struct {
	KustomizePluginData *KustomizePluginData
	OtherResources      []map[string]any
}

// tryParseKustomizePluginDataResource attempts to parse a document as KustomizePluginData.
// Returns the parsed KustomizePluginData and nil error if successful.
// Returns nil and nil if the document is not a KustomizePluginData resource.
// Returns nil and error if the document is a KustomizePluginData resource but has invalid structure.
func tryParseKustomizePluginDataResource(doc map[string]any) (*KustomizePluginData, error) {
	// Check apiVersion
	apiVersion, ok := doc["apiVersion"].(string)
	if !ok || apiVersion != APIVersion {
		return nil, nil
	}

	// Check kind
	kind, ok := doc["kind"].(string)
	if !ok || kind != Kind {
		return nil, nil
	}

	// Parse files - this is required and must be map[string]string
	filesRaw, ok := doc["files"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("KustomizePluginData 'files' field must be a map")
	}

	files := make(map[string]string, len(filesRaw))
	for k, v := range filesRaw {
		strVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("KustomizePluginData 'files' values must be strings, got non-string value for key %q", k)
		}
		files[k] = strVal
	}

	return &KustomizePluginData{
		APIVersion: apiVersion,
		Kind:       kind,
		Files:      files,
	}, nil
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
		if len(doc) == 0 {
			continue
		}

		kpd, err := tryParseKustomizePluginDataResource(doc)
		if err != nil {
			return nil, err
		}
		if kpd != nil {
			if result.KustomizePluginData != nil {
				return nil, fmt.Errorf("multiple KustomizePluginData resources found, only one is supported")
			}
			result.KustomizePluginData = kpd
		} else {
			// Keep as generic resource
			result.OtherResources = append(result.OtherResources, doc)
		}
	}

	return result, nil
}

// MarshalResources converts resources back to YAML format
func MarshalResources(resources []map[string]any) ([]byte, error) {
	if len(resources) == 0 {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)

	for _, resource := range resources {
		if err := encoder.Encode(resource); err != nil {
			return nil, fmt.Errorf("failed to encode resource: %w", err)
		}
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	return buf.Bytes(), nil
}
