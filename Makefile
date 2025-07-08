
# Default target
all: test

# Clean generated files
clean:
	rm -f coverage.txt

# Run all tests
test:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/...

# Run all tests with coverage report
test_coverage:
	go test -v -race -coverprofile=coverage.txt -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/...

# Run contract tests
test_contract:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/ -run "^Test.*Contract|^Test.*Config|^Test.*Cache|^Test.*DB|^Test.*HTTP|^Test.*Log|^Test.*Module|^Test.*Observability"

# Run core tests
test_core:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/ -run "^Test.*Core|^Test.*App"

# Run integration tests
test_integration:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/ -run "^Test.*Integration"

# Run utility tests
test_utils:
	go test -v -race -coverpkg=$$(go list ./... | grep -v '/tests' | tr '\n' ',' | sed 's/,$$//') ./tests/ -run "^Test.*HumanReadable|^Test.*Convert|^Test.*Arr"
