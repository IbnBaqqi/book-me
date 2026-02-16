.PHONY: run build test clean dev

# Run the application
run:
	@go run ./cmd/server

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@air

# Build the application
build:
	@echo "Building..."
	@go build -o bin/book-me ./cmd/api
	@echo "✓ Binary built: bin/book-me"

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@rm bin/book-me
	@rm -f coverage.out
	@echo "✓ Cleaned"

# Generate sqlc code
sqlc:
	@sqlc generate
	@echo "✓ Generated sqlc code"

# Format code
fmt:
	@go fmt ./...
	@echo "✓ Formatted code"

# Install dependencies
deps:
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"