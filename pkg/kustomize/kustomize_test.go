package kustomize

import (
	"strings"
	"testing"
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
commonLabels:
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

			if err == nil {
				// @todo doesn't Go have a simpler way to assert this?
				if len(k.Resources) != len(tt.wantRes) {
					t.Errorf("ParseKustomization() got %d resources, want %d", len(k.Resources), len(tt.wantRes))
					return
				}

				for i, res := range tt.wantRes {
					if k.Resources[i] != res {
						t.Errorf("ParseKustomization() resource[%d] = %v, want %v", i, k.Resources[i], res)
					}
				}
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
			// @todo we really need to stop using `commonLabels` in examples - it is a deprecated field
			"commonLabels": map[string]any{
				"app": "myapp",
			},
		},
	}

	data, err := k.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	str := string(data)

	// Should contain resources
	if !strings.Contains(str, "all.yaml") {
		t.Error("Marshal() output should contain all.yaml")
	}

	if !strings.Contains(str, "base.yaml") {
		t.Error("Marshal() output should contain base.yaml")
	}

	// Should contain other fields
	if !strings.Contains(str, "commonLabels") {
		t.Error("Marshal() output should contain commonLabels")
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
commonLabels:
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

	str := string(updated)

	// @todo we should assert the final YAML output here
	// Should preserve all original fields
	expectedFields := []string{"apiVersion", "kind", "commonLabels", "patches", "app", "myapp"}
	for _, field := range expectedFields {
		if !strings.Contains(str, field) {
			t.Errorf("Updated kustomization should preserve field: %s", field)
		}
	}

	// Should contain both original and new resources
	if !strings.Contains(str, "base.yaml") {
		t.Error("Updated kustomization should preserve base.yaml")
	}

	if !strings.Contains(str, "all.yaml") {
		t.Error("Updated kustomization should contain all.yaml")
	}
}
