
# Default target
all: test-contract test-core

# Run all tests
test-contract:
	go test -v ./tests/contract/...

test-core:
	go test -v ./tests/core/...