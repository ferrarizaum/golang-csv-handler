.PHONY: build build-local build-lambda test clean help

# Build both local and lambda binaries
build: build-local build-lambda

# Build local CLI application
build-local:
	@echo "Building local CLI..."
	@go build -o bin/csv-cleaner cmd/local/main.go
	@echo "✓ Local binary created: bin/csv-cleaner"

# Build Lambda function
build-lambda:
	@echo "Building Lambda function..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o lambda/bootstrap cmd/lambda/main.go
	@cd lambda && zip -q function.zip bootstrap
	@echo "✓ Lambda package created: lambda/function.zip"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run tests with coverage report
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f bin/csv-cleaner
	@rm -f lambda/bootstrap
	@rm -f lambda/function.zip
	@rm -f coverage.out
	@echo "✓ Clean complete"

# Run local example
example:
	@echo "Running example..."
	@go run cmd/local/main.go -input test-data-dirty.csv -output test-data-clean.csv

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build both local and lambda binaries"
	@echo "  build-local   - Build local CLI application"
	@echo "  build-lambda  - Build Lambda deployment package"
	@echo "  test          - Run tests with coverage"
	@echo "  test-coverage - Run tests and open coverage report in browser"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"
	@echo "  tidy          - Tidy dependencies"
	@echo "  clean         - Remove build artifacts"
	@echo "  example       - Run with test data"
	@echo "  help          - Show this help message"
