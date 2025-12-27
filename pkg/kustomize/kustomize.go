package kustomize

import (
	"bytes"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// Kustomization represents a kustomization.yaml file structure
type Kustomization struct {
	Resources []string `yaml:"resources,omitempty"`
	// We only care about the resources field for now
	// Other fields are preserved as-is using RawContent
	RawContent map[string]any
}

// ParseKustomization parses a kustomization.yaml file
func ParseKustomization(data []byte) (*Kustomization, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse kustomization.yaml: %w", err)
	}

	k := &Kustomization{
		RawContent: raw,
	}

	// Extract resources if present
	// @todo can't Go/Yaml parser do all of this for us?
	if resourcesRaw, ok := raw["resources"]; ok {
		if resourcesList, ok := resourcesRaw.([]any); ok {
			k.Resources = make([]string, 0, len(resourcesList))
			for _, r := range resourcesList {
				if s, ok := r.(string); ok {
					k.Resources = append(k.Resources, s)
				}
			}
		}
	}

	return k, nil
}

// AddResource adds a resource to the kustomization if not already present
func (k *Kustomization) AddResource(resource string) bool {
	if slices.Contains(k.Resources, resource) {
		return false // Already present
	}

	k.Resources = append(k.Resources, resource)
	k.RawContent["resources"] = k.Resources
	return true
}

// Marshal converts the kustomization back to YAML
func (k *Kustomization) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(k.RawContent); err != nil {
		return nil, fmt.Errorf("failed to marshal kustomization: %w", err)
	}

	encoder.Close()
	return buf.Bytes(), nil
}

// EnsureAllYamlInKustomization reads kustomization.yaml, ensures all.yaml is in resources,
// and returns the updated content if changes were made
func EnsureAllYamlInKustomization(kustomizationContent []byte) (updated []byte, changed bool, err error) {
	k, err := ParseKustomization(kustomizationContent)
	if err != nil {
		return nil, false, err
	}

	changed = k.AddResource("all.yaml")

	updated, err = k.Marshal()
	if err != nil {
		return nil, false, err
	}

	return updated, changed, nil
}
