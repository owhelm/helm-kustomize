package kustomize

import (
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestParseKustomization(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRes []string
		wantErr bool
	}{
		{
			name: "with resources",
			input: `resources:
- all.yaml
- deployment.yaml
patches:
- path: patch.yaml
`,
			wantRes: []string{"all.yaml", "deployment.yaml"},
			wantErr: false,
		},
		{
			name: "without resources",
			input: `patches:
- path: patch.yaml
`,
			wantRes: nil,
			wantErr: false,
		},
		{
			name:    "empty kustomization",
			input:   `{}`,
			wantRes: nil,
			wantErr: false,
		},
		{
			name: "resources with other fields",
			input: `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- base.yaml
labels:
- includeSelectors: true
  includeTemplates: true
  pairs:
    app: myapp
`,
			wantRes: []string{"base.yaml"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := ParseKustomization([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKustomization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !slices.Equal(k.Resources, tt.wantRes) {
				t.Errorf("ParseKustomization() resources = %v, want %v", k.Resources, tt.wantRes)
			}
		})
	}
}

func TestParseKustomization_InvalidYAML(t *testing.T) {
	input := `this is not: valid: yaml: structure
  bad indentation
`
	_, err := ParseKustomization([]byte(input))
	if err == nil {
		t.Fatal("ParseKustomization() should return error for invalid YAML")
	}
}

func TestParseKustomization_ResourcesNotArray(t *testing.T) {
	input := `resources: "not an array"`
	_, err := ParseKustomization([]byte(input))
	if err == nil {
		t.Fatal("ParseKustomization() should return error when resources is not an array")
	}
	if !strings.Contains(err.Error(), "resources field must be an array") {
		t.Errorf("Error should mention resources field must be an array, got: %v", err)
	}
}

func TestParseKustomization_ResourceItemNotString(t *testing.T) {
	input := `resources:
  - all.yaml
  - 123
  - base.yaml
`
	_, err := ParseKustomization([]byte(input))
	if err == nil {
		t.Fatal("ParseKustomization() should return error when resource item is not a string")
	}
	if !strings.Contains(err.Error(), "must be a string") {
		t.Errorf("Error should mention resource must be a string, got: %v", err)
	}
}

func TestKustomization_AddResource(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		add          string
		wantChanged  bool
		wantResAfter []string
	}{
		{
			name:         "add to empty list",
			initial:      []string{},
			add:          "all.yaml",
			wantChanged:  true,
			wantResAfter: []string{"all.yaml"},
		},
		{
			name:         "add to existing list",
			initial:      []string{"base.yaml"},
			add:          "all.yaml",
			wantChanged:  true,
			wantResAfter: []string{"base.yaml", "all.yaml"},
		},
		{
			name:         "add duplicate",
			initial:      []string{"all.yaml", "base.yaml"},
			add:          "all.yaml",
			wantChanged:  false,
			wantResAfter: []string{"all.yaml", "base.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kustomization{
				Resources:  tt.initial,
				RawContent: map[string]any{"resources": tt.initial},
			}

			changed := k.AddResource(tt.add)

			if changed != tt.wantChanged {
				t.Errorf("AddResource() changed = %v, want %v", changed, tt.wantChanged)
			}

			if len(k.Resources) != len(tt.wantResAfter) {
				t.Errorf("AddResource() resulted in %d resources, want %d", len(k.Resources), len(tt.wantResAfter))
				return
			}

			for i, res := range tt.wantResAfter {
				if k.Resources[i] != res {
					t.Errorf("AddResource() resource[%d] = %v, want %v", i, k.Resources[i], res)
				}
			}
		})
	}
}

func TestKustomization_Marshal(t *testing.T) {
	k := &Kustomization{
		Resources: []string{"all.yaml", "base.yaml"},
		RawContent: map[string]any{
			"resources": []string{"all.yaml", "base.yaml"},
			"labels": []map[string]any{
				{
					"pairs": map[string]string{
						"app": "myapp",
					},
				},
			},
		},
	}

	data, err := k.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	want := `labels:
    - pairs:
        app: myapp
resources:
    - all.yaml
    - base.yaml
`

	var got, expected map[string]any
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}
	if err := yaml.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected: %v", err)
	}

	gotYAML, _ := yaml.Marshal(got)
	expectedYAML, _ := yaml.Marshal(expected)

	if string(gotYAML) != string(expectedYAML) {
		t.Errorf("Marshal() output =\n%s\nwant =\n%s", string(gotYAML), string(expectedYAML))
	}
}

func TestEnsureAllYamlInKustomization(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantChanged bool
	}{
		{
			name: "all.yaml not present",
			input: `resources:
- base.yaml
`,
			wantChanged: true,
		},
		{
			name: "all.yaml already present",
			input: `resources:
- all.yaml
- base.yaml
`,
			wantChanged: false,
		},
		{
			name: "no resources field",
			input: `patches:
- path: patch.yaml
`,
			wantChanged: true,
		},
		{
			name:        "empty kustomization",
			input:       `{}`,
			wantChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, changed, err := EnsureAllYamlInKustomization([]byte(tt.input))
			if err != nil {
				t.Fatalf("EnsureAllYamlInKustomization() error = %v, want nil", err)
			}

			if changed != tt.wantChanged {
				t.Errorf("EnsureAllYamlInKustomization() changed = %v, want %v", changed, tt.wantChanged)
			}

			// Updated output should always contain all.yaml
			if !strings.Contains(string(updated), "all.yaml") {
				t.Error("Updated kustomization should contain all.yaml")
			}

			// Verify we can parse the updated output
			_, err = ParseKustomization(updated)
			if err != nil {
				t.Errorf("Updated kustomization is not valid YAML: %v", err)
			}
		})
	}
}

func TestEnsureAllYamlInKustomization_PreservesOtherFields(t *testing.T) {
	input := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- base.yaml
labels:
- includeSelectors: true
  includeTemplates: true
  pairs:
    app: myapp
    version: v1
patches:
- path: patch.yaml
`

	updated, changed, err := EnsureAllYamlInKustomization([]byte(input))
	if err != nil {
		t.Fatalf("EnsureAllYamlInKustomization() error = %v", err)
	}

	if !changed {
		t.Error("Expected kustomization to be changed")
	}

	want := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
    - base.yaml
    - all.yaml
labels:
    - includeSelectors: true
      includeTemplates: true
      pairs:
        app: myapp
        version: v1
patches:
    - path: patch.yaml
`

	var got, expected map[string]any
	if err := yaml.Unmarshal(updated, &got); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}
	if err := yaml.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected: %v", err)
	}

	gotYAML, _ := yaml.Marshal(got)
	expectedYAML, _ := yaml.Marshal(expected)

	if string(gotYAML) != string(expectedYAML) {
		t.Errorf("EnsureAllYamlInKustomization() output =\n%s\nwant =\n%s", string(gotYAML), string(expectedYAML))
	}
}

func TestEnsureAllYamlInKustomization_ParseError(t *testing.T) {
	// Test that EnsureAllYamlInKustomization returns error when ParseKustomization fails
	input := `resources: "not an array"`
	_, _, err := EnsureAllYamlInKustomization([]byte(input))
	if err == nil {
		t.Fatal("EnsureAllYamlInKustomization() should return error when ParseKustomization fails")
	}
}

func TestBuild_Error(t *testing.T) {
	// Test Build with an invalid/non-existent directory
	_, err := Build("/nonexistent/directory/that/does/not/exist")
	if err == nil {
		t.Fatal("Build() should return error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "kubectl kustomize failed") {
		t.Errorf("Error should mention kubectl kustomize failed, got: %v", err)
	}
}
