#!/bin/bash
set -euo pipefail

# Emulator script for the Helm Kustomize Plugin
# This simulates the plugin's behavior using bash and yq
# Usage: helm template ./simple-app | ./emulator.sh

# Create temporary directory and ensure cleanup
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

# Read all input from stdin
input=$(cat)

# Extract the KustomizeFiles resource
kustomize_files=$(echo "$input" | yq eval-all 'select(.kind == "KustomizeFiles" and .apiVersion == "helm.kustomize.plugin/v1alpha1")' -)

# Check if KustomizeFiles resource exists
if [ -z "$kustomize_files" ] || [ "$kustomize_files" = "null" ]; then
  # No KustomizeFiles resource found, just pass through the input
  echo "$input"
  exit 0
fi

# Get list of file paths from the files map
file_paths=$(echo "$kustomize_files" | yq eval '.files | keys | .[]' -)

# Extract and create each file
while IFS= read -r filepath; do
  # Skip empty lines
  [ -z "$filepath" ] && continue

  # Get the content for this specific file
  content=$(echo "$kustomize_files" | yq eval ".files[\"$filepath\"]" -)

  # Create directory structure if needed
  file_dir=$(dirname "$filepath")
  if [ "$file_dir" != "." ]; then
    mkdir -p "$tmpdir/$file_dir"
  fi

  # Write file content
  echo "$content" > "$tmpdir/$filepath"
done <<< "$file_paths"

# Write all non-KustomizeFiles resources to all.yaml
echo "$input" | yq eval-all 'select(.kind != "KustomizeFiles" or .apiVersion != "helm.kustomize.plugin/v1alpha1")' - > "$tmpdir/all.yaml"

# Check if kustomization.yaml exists and update it if needed
if [ -f "$tmpdir/kustomization.yaml" ]; then
  # Check if all.yaml is already in resources array
  has_all=$(yq eval '.resources // [] | contains(["all.yaml"])' "$tmpdir/kustomization.yaml")

  if [ "$has_all" != "true" ]; then
    # Add all.yaml to resources
    yq eval '.resources += ["all.yaml"]' -i "$tmpdir/kustomization.yaml"
  fi
fi

# Run kustomize build and output to stdout
# Try standalone kustomize first, fall back to kubectl kustomize
if command -v kustomize &> /dev/null; then
  kustomize build "$tmpdir"
elif command -v kubectl &> /dev/null; then
  kubectl kustomize "$tmpdir"
else
  echo "Error: neither kustomize nor kubectl found" >&2
  exit 1
fi