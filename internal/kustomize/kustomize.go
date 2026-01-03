package kustomize

import (
	"bytes"
	"fmt"
	"os/exec"
	"slices"

	"go.yaml.in/yaml/v4"
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
	if resourcesRaw, ok := raw["resources"]; ok {
		resourcesList, ok := resourcesRaw.([]any)
		if !ok {
			return nil, fmt.Errorf("resources field must be an array")
		}
		k.Resources = make([]string, 0, len(resourcesList))
		for i, r := range resourcesList {
			s, ok := r.(string)
			if !ok {
				return nil, fmt.Errorf("resources[%d] must be a string, got %T", i, r)
			}
			k.Resources = append(k.Resources, s)
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

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

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

// Build runs kubectl kustomize on the given directory and returns the output
func Build(dir string) ([]byte, error) {
	cmd := exec.Command("kubectl", "kustomize", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl kustomize failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}
