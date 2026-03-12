.PHONY: help lint lint-fix format test run run-verbose install-tools hook-install hook-install-light hook-uninstall hook-test

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required Go tools
	@echo "📦 Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "📦 Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Tools installed"

lint: ## Run linter to check code quality
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run ./...

lint-fix: ## Run linter and auto-fix issues
	@echo "🔧 Running golangci-lint with auto-fix..."
	@golangci-lint run --fix ./...

format: ## Format all Go code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@$$HOME/go/bin/goimports -w . 2>/dev/null || goimports -w . || true

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

run: ## Run the server
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

run-verbose: ## Run the server with verbose logging
	@echo "🚀 Starting server with verbose logging..."
	@LOG_VERBOSE=true go run cmd/server/main.go

build: ## Build the binary
	@echo "🔨 Building..."
	@go build -o bin/leona-scanner cmd/server/main.go
	@echo "✅ Binary: bin/leona-scanner"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf bin/ coverage.out coverage.html
	@echo "✅ Cleaned"

check: format lint test ## Run format, lint, and tests (full check)
	@echo "✅ All checks passed!"

hook-install: ## Install pre-commit hook (format + lint + tests)
	@echo "🔗 Installing pre-commit hook..."
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed (full checks)"
	@echo "💡 To use light version: make hook-install-light"

hook-install-light: ## Install lightweight pre-commit hook (format + lint only)
	@echo "🔗 Installing lightweight pre-commit hook..."
	@cp scripts/pre-commit-light .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Lightweight pre-commit hook installed"
	@echo "⚠️  Tests will NOT run automatically"

hook-uninstall: ## Remove pre-commit hook
	@echo "🗑️  Removing pre-commit hook..."
	@rm -f .git/hooks/pre-commit
	@echo "✅ Pre-commit hook removed"

hook-test: ## Test the pre-commit hook without committing
	@echo "🧪 Testing pre-commit hook..."
	@.git/hooks/pre-commit
