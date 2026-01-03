# Simple App Example

This example demonstrates how to use the Helm Kustomize Plugin with a basic application.

## What This Example Does

This chart deploys a simple nginx application and uses kustomize to apply transformations:

1. **Base Resources**: The chart includes a Deployment and Service
2. **Kustomize Patches**: The embedded kustomization applies:
   - Increases replica count from 1 to 3
   - Adds an `environment: production` label

## Chart Structure

```
simple-app/
├── Chart.yaml
├── values.yaml
├── kustomization/                # Kustomize files directory
│   ├── kustomization.yaml        # Main kustomization config
│   └── patches/
│       ├── deployment-replicas.yaml
│       └── add-label.yaml
├── expected-output/              # Expected final output after plugin processing
│   ├── deployment.yaml           # Deployment with patches applied
│   └── service.yaml              # Service (unchanged)
└── templates/
    ├── deployment.yaml           # Basic nginx deployment
    ├── service.yaml              # ClusterIP service
    └── kustomize-files.yaml      # Template using .Files to embed kustomization files
```

## How It Works

1. When Helm renders the chart, it produces:
   - A Deployment with 1 replica (from values.yaml)
   - A Service
   - A KustomizePluginData resource containing the kustomization

2. The plugin processes the output:
   - Detects the KustomizePluginData resource
   - Extracts the embedded files to a temporary directory
   - Writes all other resources to `all.yaml`
   - Updates `kustomization.yaml` to reference `all.yaml`
   - Runs `kustomize build` on the temporary directory
   - Returns the transformed output to Helm

3. The final result:
   - Deployment has 3 replicas (patched by kustomize)
   - Deployment has additional `environment: production` label
   - No KustomizePluginData resource in final output

## Expected Output

The `expected-output/` directory contains the final resources as they should appear after the plugin processes them. This shows:

- **deployment.yaml**: The deployment with all kustomize patches applied
  - `replicas: 3` (increased from 1)
  - `environment: production` label added to metadata
- **service.yaml**: The service unchanged (no patches target it)
