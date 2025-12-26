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
└── templates/
    ├── deployment.yaml           # Basic nginx deployment
    ├── service.yaml              # ClusterIP service
    └── kustomize-files.yaml      # Embedded kustomization files
```

## How It Works

1. When Helm renders the chart, it produces:
   - A Deployment with 1 replica (from values.yaml)
   - A Service
   - A KustomizeFiles resource containing the kustomization

2. The plugin processes the output:
   - Detects the KustomizeFiles resource
   - Extracts the embedded files to a temporary directory
   - Writes all other resources to `all.yaml`
   - Updates `kustomization.yaml` to reference `all.yaml`
   - Runs `kustomize build` on the temporary directory
   - Returns the transformed output to Helm

3. The final result:
   - Deployment has 3 replicas (patched by kustomize)
   - Deployment has additional `environment: production` label
   - No KustomizeFiles resource in final output