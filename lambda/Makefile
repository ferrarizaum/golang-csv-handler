# Makefile simplifies building and deploying the Lambda function
# Make is a build automation tool that runs commands defined in this file

# Detect OS
ifeq ($(OS),Windows_NT)
    # Windows
    BUILD_CMD = powershell -ExecutionPolicy Bypass -File build.ps1
else
    # Unix-like (Linux, macOS, Git Bash, WSL)
    # Build from current directory (lambda has its own go.mod)
    BUILD_CMD = GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap .
endif

# Build the Lambda function for Linux (required for AWS Lambda)
# Lambda runs on Amazon Linux, so we need to compile for Linux
# GOOS=linux sets the target operating system
# GOARCH=amd64 sets the target architecture (64-bit)
# CGO_ENABLED=0 creates a static binary (no external dependencies)
.PHONY: build
build:
	@echo "Building Lambda function..."
	$(BUILD_CMD)
	@echo "Build complete! Binary: bootstrap"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f bootstrap
	@echo "Clean complete!"

# Run tests (if you add tests later)
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	@echo "Dependencies installed!"

