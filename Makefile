.PHONY: test test_coverage clean lint fmt vet

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

# Run unit tests only (excludes integration tests)
test:
	$(GOTEST) -v -race $(TEST_PACKAGES)

# Run integration tests (requires external services like Redis)
test_integration:
	$(GOTEST) -v -race -tags=integration $(TEST_PACKAGES)

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