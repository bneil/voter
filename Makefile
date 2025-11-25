.PHONY: build clean test bin-dir

# Binary name and path
BINARY_NAME=voter
BINARY_PATH=./bin/$(BINARY_NAME)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) mod download

# Main package path
MAIN_PACKAGE=./cmd/voter

# Build the binary
build: bin-dir
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PACKAGE)

# Create bin directory
bin-dir:
	mkdir -p ./bin

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf ./bin

# Run tests
test:
	$(GOTEST) ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -cover ./...

# Download dependencies
deps:
	$(GOGET)

# Install the binary to $GOPATH/bin
install: build
	$(GOCMD) install $(MAIN_PACKAGE)

# Run the binary (requires build first)
run: build
	./bin/$(BINARY_NAME)

# Default target
all: build