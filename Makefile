.PHONY: test test_coverage clean lint fmt vet test_services_up test_services_down test_integration test_integration_full

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Test parameters
TEST_PACKAGES=./...
COVERAGE_FILE=coverage.txt
COVERAGE_MODE=atomic

# Default target
all: test

# Test tagging convention
# ------------------------
# A test file carries `//go:build integration` if and only if it REQUIRES an
# external service (Redis, MongoDB, Postgres). Everything else belongs in the
# default lane, whatever it is named.
#
# Tests that can use a service but do not need one — such as core/lock — stay in
# the default lane and skip themselves via a single package-level probe. That way
# they run automatically wherever the service happens to be available, without
# costing anything when it is not. Point them elsewhere with GOE_TEST_REDIS_ADDR.

# Run unit tests only (excludes integration tests)
test:
	$(GOTEST) -v -race $(TEST_PACKAGES)

# Start/stop the service stack the integration tests need.
# MongoDB runs as a single-node replica set because the migrator uses
# transactions, which standalone MongoDB rejects.
test_services_up:
	docker compose -f docker-compose.test.yml up -d --wait

test_services_down:
	docker compose -f docker-compose.test.yml down -v

# Run integration tests (requires the service stack above to be running)
#
# -p 1 serializes packages. They share one Redis and one MongoDB, so running
# them concurrently — which is the default — makes them compete for the same
# single-threaded Redis while the race detector is already slowing everything
# down. That produced roughly one spurious failure every nine full-suite runs.
test_integration:
	$(GOTEST) -v -race -p 1 -tags=integration $(TEST_PACKAGES)

# Start services, run integration tests, tear services down again
test_integration_full: test_services_up
	$(GOTEST) -race -p 1 -tags=integration $(TEST_PACKAGES); \
	status=$$?; \
	$(MAKE) test_services_down; \
	exit $$status

# Run all tests (unit + integration)
test_all:
	$(GOTEST) -v -race -tags=integration $(TEST_PACKAGES)

# Run tests with coverage (unit tests only)
test_coverage:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=$(COVERAGE_MODE) $(TEST_PACKAGES)

# Run tests with coverage including integration tests
test_coverage_all:
	$(GOTEST) -v -race -tags=integration -coverprofile=$(COVERAGE_FILE) -covermode=$(COVERAGE_MODE) $(TEST_PACKAGES)

# Run tests with coverage and generate HTML report
test_coverage_html: test_coverage
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(COVERAGE_FILE) coverage.html

# Format code
fmt:
	$(GOFMT) -s -w .

# Vet code
vet:
	$(GOCMD) vet $(TEST_PACKAGES)

# Run linter (if golangci-lint is available)
lint:
	@which $(GOLINT) > /dev/null 2>&1 || (echo "golangci-lint not found, skipping lint" && exit 0)
	$(GOLINT) run

# Download dependencies
deps:
	$(GOMOD) download

# Verify dependencies
verify:
	$(GOMOD) verify

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Check for vulnerabilities (if govulncheck is available)
vuln:
	@which govulncheck > /dev/null 2>&1 || (echo "govulncheck not found, skipping vulnerability check" && exit 0)
	govulncheck ./...

# Run all checks
check: fmt vet lint test_coverage vuln

# Development workflow
dev: deps fmt vet test

# CI workflow
ci: verify fmt vet test_coverage

help:
	@echo "Available targets:"
	@echo "  test           - Run unit tests only (excludes integration tests)"
	@echo "  test_integration - Run integration tests (requires external services)"
	@echo "  test_all       - Run all tests (unit + integration)"
	@echo "  test_coverage  - Run unit tests with coverage"
	@echo "  test_coverage_all - Run all tests with coverage"
	@echo "  test_coverage_html - Run tests with coverage and generate HTML report"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  lint           - Run linter"
	@echo "  deps           - Download dependencies"
	@echo "  verify         - Verify dependencies"
	@echo "  tidy           - Tidy dependencies"
	@echo "  vuln           - Check for vulnerabilities"
	@echo "  check          - Run all checks"
	@echo "  dev            - Development workflow"
	@echo "  ci             - CI workflow"
	@echo "  help           - Show this help"