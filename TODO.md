# TODO List - Helm Kustomize Plugin

## Error Handling & Edge Cases

- [ ] Handle invalid or malformed special resources with helpful error messages
- [ ] Improve kustomize execution error messages
- [ ] Add timeout handling for kustomize operations
- [ ] Test error cases: empty KustomizePluginData, missing kustomization.yaml
- [ ] Test error cases: kustomize build failures with clear output
- [ ] Add structured error messages with context
- [ ] Validate KustomizePluginData structure before processing
- [ ] Add error recovery suggestions in messages
- [ ] Test all error paths

## Testing

- [ ] Makefile: use custom `HELM_PLUGINS` path when testing
- [ ] Makefile: add shortcut to verify the example simple-app
- [ ] Add test case: chart without KustomizePluginData (pass-through)
- [ ] Add test case: malformed KustomizePluginData resource
- [ ] Add test case: nested directory structures
- [ ] 100% coverage
- [ ] Add more edge case tests
- [ ] Add performance tests for large charts
- [ ] Test cleanup on error conditions
- [ ] Add test for concurrent plugin usage
- [ ] Document test strategy

## Documentation

- [ ] Write README with installation instructions
  - [ ] Prerequisites (Helm v4, kubectl)
  - [ ] Installation steps
  - [ ] Basic usage example
  - [ ] Link to examples
- [ ] Add development guide
  - [ ] How to build from source
  - [ ] How to run tests
  - [ ] How to contribute

## Examples & Samples

- [ ] Add example: multiple patches
- [ ] Add example: labels (not commonLabels, which is deprecated!) usage
- [ ] Add example: image transformations
- [ ] Add example: modify all resources of certain `Kind` with annotations

## Build & Distribution

- [ ] Set up CI/CD pipeline
- [ ] Create release automation
- [ ] Add version management

## Future Enhancements

- [ ] Support for multiple kustomization files
- [ ] Configurable resource naming (alternative to `all.yaml`)
- [ ] Performance optimization for large charts
