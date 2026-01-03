#!/bin/bash
set -euo pipefail

# Integration test for helm-kustomize-plugin
# Tests the plugin against the simple-app example

echo "Running integration test..."

# Ensure plugin is installed
if ! helm plugin list | grep -q kustomize; then
  echo "Error: kustomize plugin not installed"
  echo "Run 'make install' first"
  exit 1
fi

# Run helm template with the plugin
echo "Testing simple-app example..."
OUTPUT=$(helm template examples/simple-app --post-renderer kustomize)

# Check for expected transformations
FAILED=0

# Check 1: Deployment should have 3 replicas (patched from 1)
if echo "$OUTPUT" | grep -q "replicas: 3"; then
  echo "✓ Deployment replicas patched to 3"
else
  echo "✗ FAILED: Deployment replicas not set to 3"
  FAILED=1
fi

# Check 2: Deployment should have environment: production label
if echo "$OUTPUT" | grep -q "environment: production"; then
  echo "✓ Environment label added"
else
  echo "✗ FAILED: Environment label not found"
  FAILED=1
fi

# Check 3: KustomizePluginData resource should NOT be in output
if echo "$OUTPUT" | grep -q "kind: KustomizePluginData"; then
  echo "✗ FAILED: KustomizePluginData resource found in output (should be removed)"
  FAILED=1
else
  echo "✓ KustomizePluginData resource removed from output"
fi

# Check 4: Service should still be present
if echo "$OUTPUT" | grep -q "kind: Service"; then
  echo "✓ Service resource present"
else
  echo "✗ FAILED: Service resource not found"
  FAILED=1
fi

# Check 5: Deployment should still be present
if echo "$OUTPUT" | grep -q "kind: Deployment"; then
  echo "✓ Deployment resource present"
else
  echo "✗ FAILED: Deployment resource not found"
  FAILED=1
fi

echo ""
if [ $FAILED -eq 0 ]; then
  echo "✅ All integration tests passed!"
  exit 0
else
  echo "❌ Some tests failed"
  exit 1
fi
