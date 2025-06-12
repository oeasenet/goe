
# Default target
all: test

# Run all tests
test:
	go test -v -race -coverpkg=$(go list ./... | grep -v '/tests' | paste -sd,) ./...