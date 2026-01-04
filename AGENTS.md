# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Helm v4 post-renderer plugin that integrates Kustomize transformations into Helm chart deployments. The plugin runs as a subprocess during Helm rendering, extracting embedded Kustomize files from a special Kubernetes resource, applying them to the rendered manifests, and returning the transformed output.

## Requirements

- Helm v4 with subprocess runtime support for post-renderer plugins
- kubectl (for `kubectl kustomize` command)
- Go 1.21+ for development

## Build and Test Commands

```bash
# Build the plugin binary (automatically runs go fmt)
make build

# Format code manually
go fmt ./...

# Run unit tests (automatically runs golangci-lint)
make test
go test -v ./...

# Run linter manually
golangci-lint run

# Run a specific test
go test -v ./internal/parser -run TestParseManifests

# Run integration tests (requires Helm v4)
make test-integration

# Run all tests (unit + integration) with coverage check and HTML report generation
make test-all

# Display coverage summary (requires coverage.out from prior test run)
make coverage-report

# Install plugin into Helm
make install

# Uninstall plugin from Helm
make uninstall

# Development cycle: rebuild and reinstall
make reinstall

# Clean build artifacts and coverage files
make clean
```

## Architecture

### Data Flow

1. **Input**: Helm renders a chart to YAML manifests containing:
   - Regular Kubernetes resources
   - A special `KustomizePluginData` resource (apiVersion: `helm.plugin.kustomize/v1`, kind: `KustomizePluginData`)

2. **Processing Pipeline** (`main.go:Run()`):
   - **Parse** (`parser` package): Separates `KustomizePluginData` from other resources
   - **Extract** (`extractor` package): Creates temporary directory and extracts embedded files
   - **Generate**: Writes remaining Helm resources to `all.yaml`
   - **Patch** (`kustomize` package): Ensures `all.yaml` is referenced in `kustomization.yaml`
   - **Transform**: Runs `kubectl kustomize` on the temporary directory
   - **Output**: Returns transformed manifests to Helm

3. **Cleanup**: Temporary directory is automatically removed via defer

### Package Structure

- **`main.go`**: Entry point implementing Helm's `PostRenderer` interface. Orchestrates the entire pipeline.

- **`internal/parser`**: YAML document parsing and resource separation
  - Identifies `KustomizePluginData` resources by apiVersion/kind
  - Validates the `files` field structure (must be `map[string]string`)
  - Enforces single `KustomizePluginData` resource per chart
  - Marshals remaining resources back to YAML

- **`internal/extractor`**: Temporary filesystem management
  - Uses `os.OpenRoot()` for path-constrained file operations (security feature)
  - Creates directory structures from file paths (e.g., `patches/deployment.yaml`)
  - Handles cleanup with graceful error reporting

- **`internal/kustomize`**: Kustomization file manipulation and execution
  - Parses `kustomization.yaml` preserving all fields via `map[string]any`
  - Adds `all.yaml` to `resources` array if not present
  - Executes `kubectl kustomize` command

### Key Design Decisions

1. **YAML Parsing Strategy**: Parse once into `map[string]any` to preserve all fields, then manually extract typed fields. This avoids double-parsing overhead while maintaining round-trip fidelity.

2. **Type Conversions**: YAML unmarshaling into `map[string]any` always produces `[]any` for arrays, never typed slices like `[]string`. Manual iteration with type assertions is required.

3. **Security**: Uses `os.OpenRoot()` to constrain all file operations to the temporary directory, preventing path traversal attacks from malicious file paths.

4. **Reserved Filename**: The name `all.yaml` is reserved for Helm-rendered manifests and cannot appear in `KustomizePluginData.files`.

5. **Error Handling**: All type conversions and validations fail fast with descriptive errors rather than silently skipping invalid data.

## KustomizePluginData Resource

The special resource format that triggers kustomize processing:

```yaml
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
files:
  kustomization.yaml: |
    resources:
    - all.yaml
    patches:
    - path: patch.yaml

  patch.yaml: |
    apiVersion: apps/v1
    kind: Deployment
    # ... patch content
```

**Important constants** (`internal/parser/parser.go`):
- `APIVersion = "helm.plugin.kustomize/v1"`
- `Kind = "KustomizePluginData"`

## Testing

- Unit tests use table-driven patterns with `t.Run()` subtests
- Integration tests in `test-integration.sh` test the full plugin with actual Helm charts
- Example charts in `examples/` directory serve as test fixtures and documentation
- **YAML Output Assertions**: When testing functions that produce YAML output, assert against the final YAML string directly rather than parsing and checking individual fields. This ensures exact output validation and catches formatting issues.

## Kustomize API Preferences

- **NEVER** use deprecated `commonLabels` (deprecated in Kustomize v5.3+). Use `labels` field instead:
  ```yaml
  labels:
  - includeSelectors: true
    includeTemplates: true
    pairs:
      app: myapp
  ```
- `commonAnnotations` is NOT deprecated and can still be used