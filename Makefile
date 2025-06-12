
# Default target
all: test

# Run all tests
test:
	go test -v -race ./tests/...