.PHONY: help merge build clean install run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	@go mod download
	@echo "✅ Dependencies installed"

merge: ## Merge OpenAPI specs into consolidated file
	@echo "🔄 Merging OpenAPI specifications..."
	@go run merge_specs.go

build: ## Build the merge tool executable
	@echo "🔨 Building merge-specs executable..."
	@go build -o merge-specs merge_specs.go
	@echo "✅ Build complete: ./merge-specs"

run: build ## Build and run the merge tool
	@echo "🚀 Running merge tool..."
	@./merge-specs

clean: ## Remove generated files
	@echo "🧹 Cleaning up..."
	@rm -f merge-specs consolidated-openapi.yml
	@echo "✅ Cleaned"

validate: ## Validate the consolidated spec (requires swagger-cli)
	@echo "🔍 Validating consolidated spec..."
	@if command -v swagger-cli >/dev/null 2>&1; then \
		swagger-cli validate consolidated-openapi.yml; \
	else \
		echo "⚠️  swagger-cli not found. Install with: npm install -g @apidevtools/swagger-cli"; \
	fi
