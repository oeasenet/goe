
# Default target
all: test

# Clean generated files
clean:
	rm -f coverage.txt

# Run all tests
test: test_contract test_core test_integration

# Run all tests with coverage report
test_coverage:
	go test -v -race -coverprofile=coverage.txt -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./...

# Run test in the contact folder
test_contract:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/contract/...

# Run test in the core folder
test_core:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/core/...

# Run test in the integration folder
test_integration:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/integration/...
