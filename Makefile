
# Default target
all: test

# Run all tests
test: test_contract test_core test_integration

# Run test in the contact folder
test_contract:
	go test -v -race -coverpkg=$(go list ./... | grep -v '/tests' | paste -sd,) ./tests/contract/...

# Run test in the core folder
test_core:
	go test -v -race -coverpkg=$(go list ./... | grep -v '/tests' | paste -sd,) ./tests/core/...

# Run test in the integration folder
test_integration:
	go test -v -race -coverpkg=$(go list ./... | grep -v '/tests' | paste -sd,) ./tests/integration/...