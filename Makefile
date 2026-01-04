.PHONY: build clean test test-integration test-all install uninstall reinstall

BINARY_NAME=helm-kustomize-plugin
BUILD_DIR=.

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	go clean

test:
	go test -v ./...

test-integration: reinstall
	./test-integration.sh

test-all: test test-integration

install: build
	helm plugin install .

uninstall:
	helm plugin uninstall kustomize 2>/dev/null || true

# Development: uninstall, rebuild, and reinstall
reinstall: uninstall build install
