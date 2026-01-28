# GoDataCleaner Makefile

# Binary name
BINARY_NAME=godatacleaner

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Build flags
CGO_ENABLED=1
LDFLAGS=-ldflags "-s -w"

.PHONY: all build run test clean deps vet fmt

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/godatacleaner

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Run tests
test:
	@echo "Running tests..."
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.db

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Help
help:
	@echo "GoDataCleaner Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the application"
	@echo "  make run            Build and run the application"
	@echo "  make run ARGS=sync  Run with specific command"
	@echo "  make test           Run tests with race detector"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make clean          Remove build artifacts"
	@echo "  make deps           Download and tidy dependencies"
	@echo "  make vet            Run go vet"
	@echo "  make fmt            Format code"
	@echo "  make help           Show this help"
