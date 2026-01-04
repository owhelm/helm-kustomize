.PHONY: build clean test test-integration test-all install uninstall reinstall

BINARY_NAME=helm-kustomize-plugin
BUILD_DIR=dist

build:
	go fmt ./...
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	cp plugin.yaml $(BUILD_DIR)/

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	golangci-lint run
	go test -v ./...

test-integration: reinstall
	./test-integration.sh

test-all: test test-integration

install: build
	helm plugin install $(BUILD_DIR)

uninstall:
	helm plugin uninstall helm-kustomize 2>/dev/null || true

# Development: uninstall, rebuild, and reinstall
reinstall: uninstall build install
