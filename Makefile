BINARY_NAME=peep
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build clean test run deps

# Build the binary
build: deps
	@echo "🔨 Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f logs.db

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run the application
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Development: watch for changes and rebuild
dev:
	@echo "👀 Watching for changes..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Cross-compile for different platforms
build-all: deps
	@echo "🌍 Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# Demo: generate some sample logs and ingest them
demo: build
	@echo "🎪 Running demo..."
	@echo '{"timestamp":"2023-08-06T10:30:45Z","level":"info","message":"Application started","service":"api"}' | ./$(BINARY_NAME)
	@echo '{"timestamp":"2023-08-06T10:30:46Z","level":"error","message":"Database connection failed","service":"api"}' | ./$(BINARY_NAME)
	@echo '2023-08-06 10:30:47 WARN [web] High memory usage detected' | ./$(BINARY_NAME)

# Help
help:
	@echo "🔍 Peep - Observability for humans"
	@echo ""
	@echo "Available commands:"
	@echo "  make build     - Build the binary"
	@echo "  make deps      - Install dependencies"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make test      - Run tests"
	@echo "  make run       - Build and run"
	@echo "  make dev       - Watch for changes and rebuild"
	@echo "  make build-all - Cross-compile for all platforms"
	@echo "  make demo      - Run a quick demo"
	@echo "  make help      - Show this help"
