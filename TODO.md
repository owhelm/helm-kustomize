# TODO List - Helm Kustomize Plugin

## Project Setup

- [ ] Initialize project structure and build system
- [ ] Configure Extism WASM build environment
- [ ] Set up development dependencies (Go/Rust/language of choice)
- [ ] Create basic plugin skeleton with Helm v4 API interface

## Core Functionality

### Special Resource Detection
- [ ] Define the schema/structure for the special resource
- [ ] Implement logic to detect the special resource in chart output
- [ ] Add validation for special resource format

### File Extraction
- [ ] Implement extraction of embedded files from special resource
- [ ] Security: disallow directory traversal bugs when extracting files
- [ ] Create temporary folder management for extracted files
- [ ] Handle file permissions and directory structure preservation

### Chart Processing
- [ ] Implement removal of special resource from chart output
- [ ] Write remaining chart contents to `all.yaml`
- [ ] Implement `kustomization.yaml` update logic
  - [ ] Check if `all.yaml` is already referenced under `resources`
  - [ ] Add `all.yaml` to resources if not present
  - [ ] Preserve existing kustomization.yaml structure

### Kustomize Integration
- [ ] Implement kustomize execution against temporary folder
- [ ] Capture kustomize output
- [ ] Handle kustomize errors and validation

### Output Handling
- [ ] Format output for Helm post-renderer
- [ ] Send processed output back to Helm
- [ ] Clean up temporary files and folders

## Error Handling & Edge Cases

- [ ] Add error handling for missing special resource
- [ ] Handle invalid or malformed special resources
- [ ] Handle kustomize execution failures
- [ ] Handle file system errors (permissions, disk space, etc.)
- [ ] Add timeout handling for kustomize operations

## Testing

- [ ] Create unit tests for special resource detection
- [ ] Create unit tests for file extraction
- [ ] Create unit tests for kustomization.yaml updates
- [ ] Create integration tests with sample Helm charts
- [ ] Create end-to-end tests with actual Helm installation
- [ ] Add test fixtures and sample charts

## Documentation

- [ ] Write README with installation instructions
- [ ] Document the special resource format/schema
- [x] Create usage examples
- [ ] Document configuration options
- [ ] Add troubleshooting guide

## Examples & Samples

- [x] Create example Helm chart with embedded kustomize files
- [x] Create example kustomization.yaml
- [x] Create example overlays/patches
- [x] Document example use cases

## Build & Distribution

- [ ] Set up CI/CD pipeline
- [ ] Configure WASM build process
- [ ] Create release automation
- [ ] Add installation instructions for Helm v4

## Future Enhancements

- [ ] Support for multiple kustomization files
- [ ] Configurable resource naming (alternative to `all.yaml`)
- [ ] Dry-run mode for debugging
- [ ] Verbose logging options
- [ ] Performance optimization for large charts
