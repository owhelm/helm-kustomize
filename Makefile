.PHONY: build clean test test-integration test-all install uninstall reinstall \
        coverage-report coverage-clean

BINARY_NAME=helm-kustomize-plugin
BUILD_DIR=dist
COVERAGE_THRESHOLD=80
COVERAGE_PROFILE=coverage.out
COVERAGE_HTML=coverage.html
COVERAGE_DIR=coverage

build:
	go fmt ./...
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	cp plugin.yaml $(BUILD_DIR)/

clean: coverage-clean uninstall
	rm -rf $(BUILD_DIR)
	go clean

test:
ifndef GITHUB_ACTIONS
	golangci-lint run
endif
	go test -v -coverprofile=$(COVERAGE_PROFILE) -covermode=atomic ./...

test-integration: reinstall
	./test-integration.sh

test-all: test test-integration
	@echo "Checking coverage threshold (${COVERAGE_THRESHOLD}%)..."
	@bash -c 'coverage=$$(go tool cover -func=$(COVERAGE_PROFILE) | tail -1 | awk "{print int(\$$3)}"); \
	if [ $$coverage -lt $(COVERAGE_THRESHOLD) ]; then \
	  echo "Coverage $$coverage% is below threshold $(COVERAGE_THRESHOLD)%"; \
	  exit 1; \
	else \
	  echo "Coverage $$coverage% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi'
	@mkdir -p $(COVERAGE_DIR)
	@go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_DIR)/$(COVERAGE_HTML)
	@echo "Coverage HTML report generated at $(COVERAGE_DIR)/$(COVERAGE_HTML)"

coverage-report:
	@echo "=== Coverage Summary ==="
	@go tool cover -func=$(COVERAGE_PROFILE) | tail -1 | awk '{print "Overall Coverage: " $$3}'
	@echo ""
	@echo "=== Coverage by Package ==="
	@go tool cover -func=$(COVERAGE_PROFILE) | grep -E "^github.com/owhelm" | grep -v "total"

coverage-clean:
	rm -rf $(COVERAGE_DIR) $(COVERAGE_PROFILE) $(COVERAGE_HTML) helm-kustomize-*.tgz

install: build
	helm plugin install $(BUILD_DIR)

uninstall:
	helm plugin uninstall helm-kustomize 2>/dev/null || true

# Development: uninstall, rebuild, and reinstall
reinstall: uninstall build install
