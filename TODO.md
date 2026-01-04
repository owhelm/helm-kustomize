# TODO List - Helm Kustomize Plugin

## Testing

- [ ] Makefile: use custom `HELM_PLUGINS` path when testing
- [ ] 100% coverage
- [ ] Test in multiple kubectl versions (broken due to https://github.com/Azure/setup-kubectl/issues/88)

## Documentation

- [ ] Write README with installation instructions
  - [ ] Prerequisites (Helm v4, kubectl)
  - [ ] Installation steps
  - [ ] Basic usage example
  - [ ] Link to examples

## Build & Distribution

- [x] Set up CI/CD pipeline
- [ ] Create release automation
- [ ] Add version management
- [ ] Set up renovate for ci.yaml

## Future Enhancements

- [ ] Support for multiple kustomization files
- [ ] Configurable resource naming (alternative to `all.yaml`)
- [ ] Performance optimization for large charts
- [ ] Support helm v3
