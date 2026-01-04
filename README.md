# experiment-helm-kustomize-plugin

This is an experiment to build a Kustomize plugin for Helm.

## Requirements

- Helm v4 (with subprocess runtime support for post-renderer plugins)
- kubectl (for `kubectl kustomize` command)

## Design

- The plugin uses the Helm v4 plugin API with subprocess runtime
- It's a post-renderer plugin:
  - It expects the Helm chart to contain a special resource, which includes all the relevant files embedded inside of it
  - If it finds the special resource inside the chart
    - it extracts all the files contained in the special resource into a temporary folder
    - it removes the special resource from the chart output
    - it outputs the entire remaining contents of the chart into the `all.yaml` file
    - it updates the `kustomization.yaml` to reference the `all.yaml` under `resources` if it's not already referenced
    - it runs `kubectl kustomize` against the temporary folder and captures the output
    - it sends the output back to Helm

## Special Resource Format

The plugin uses a custom Kubernetes resource to embed kustomize files within a Helm chart. This resource is detected during post-rendering and used to apply kustomize transformations.

### Schema

```yaml
apiVersion: helm.kustomize.plugin/v1alpha1
kind: KustomizePluginData
metadata:
  name: kustomize-files
files:
  kustomization.yaml: |
    resources:
    - all.yaml

    patches:
    - path: patch.yaml

  patch.yaml: |
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-app
    spec:
      replicas: 3

  # Additional files can be included as needed
  # Supports any file structure required by kustomize
```

### Field Descriptions

- **apiVersion**: Must be `helm.kustomize.plugin/v1alpha1`
- **kind**: Must be `KustomizePluginData`
- **metadata.name**: Identifier for the resource (can be any valid Kubernetes name)
- **files**: A map where keys are file paths and values are file contents
  - File paths can include directories (e.g., `overlays/production/patch.yaml`)
  - Contents are embedded as strings (potentially using YAML multi-line)
  - At minimum, should include a `kustomization.yaml` file

### File Structure

The `files` map supports nested directory structures by using path separators in the keys:

```yaml
files:
  kustomization.yaml: |
    # Main kustomization file

  patches/deployment.yaml: |
    # Patch file in patches subdirectory

  overlays/production/kustomization.yaml: |
    # Overlay-specific kustomization
```

When extracted, the plugin will create the appropriate directory structure in the temporary folder.

### Requirements

1. The resource must have `apiVersion: helm.kustomize.plugin/v1alpha1` and `kind: KustomizePluginData`
2. At least one file must be specified in the `files` map
3. A `kustomization.yaml` file should be present in the root (though kustomize can work with nested kustomizations)
4. File contents must be valid YAML or appropriate format for kustomize processing

### Notes

- This resource is automatically removed from the final chart output after processing
- Multiple `KustomizePluginData` resources in a single chart are not currently supported
- The resource is processed before the final render, so kustomize transformations are applied to all chart resources

## Use Cases

Some of the use cases below are generic kustomize features, where it excels against Helm. 

### 1. Cross-Cutting Concerns with Common Labels

Add labels, annotations, or other metadata across all resources:

```yaml
files:
  kustomization.yaml: |
    resources:
    - all.yaml

    labels:
    - includeSelectors: true
      includeTemplates: true
      pairs:
        team: platform
        cost-center: engineering
        compliance: pci

    commonAnnotations:
      managed-by: platform-team
      security-scan: enabled
```

**Why use this**: Instead of templating these into every resource, kustomize applies them uniformly. Easy to add/remove without touching individual templates.

### 2. Image Tag Management and Digests

Override image tags or add digest pinning for security:

```yaml
files:
  kustomization.yaml: |
    resources:
    - all.yaml

    images:
    - name: nginx
      newTag: 1.21.6
      digest: sha256:a5b8b7a...
    - name: redis
      newName: custom-registry.io/redis
      newTag: 7.0-alpine
```

**Why use this**: Centralizes image management and enables digest pinning for supply chain security without modifying deployment templates.

### 3. Strategic Merge Patches for Complex Changes

Apply sophisticated patches that Helm templating would make unwieldy:

```yaml
files:
  kustomization.yaml: |
    resources:
    - all.yaml
    patches:
    - path: add-sidecar.yaml

  add-sidecar.yaml: |
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-app
    spec:
      template:
        spec:
          containers:
          - name: istio-proxy
            image: istio/proxyv2:1.18.0
            ports:
            - containerPort: 15090
              name: http-envoy-prom
```

**Why use this**: Adding sidecars, init containers, or volume mounts via Helm templating can be complex and error-prone. Kustomize patches are cleaner.

### 4. JSON 6902 Patches for Precise Modifications

Make surgical changes to specific fields:

```yaml
files:
  kustomization.yaml: |
    resources:
    - all.yaml
    patches:
    - target:
        kind: Deployment
        name: my-app
      patch: |-
        - op: replace
          path: /spec/strategy/type
          value: Recreate
        - op: add
          path: /spec/template/spec/securityContext
          value:
            runAsNonRoot: true
            runAsUser: 1000
```

**Why use this**: When you need exact control over specific fields, JSON patches are more precise than strategic merges or Helm templating.

### 5. Combining Helm's Strengths with Kustomize

Use Helm for:
- Package management and versioning
- Initial resource templating
- Dependency management

Use Kustomize for:
- Final transformations and patches
- Cross-cutting concerns
- Image management

## Examples

See the [`examples/`](./examples) directory for complete working examples:

- **[simple-app](./examples/simple-app)**: Basic application with kustomize patches demonstrating replica count and label modifications
