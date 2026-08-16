.PHONY: help check test test-integration test-coverage test-verbose clean run-api run-processor setup verify

TRANSCODER_REGION ?= us-central1
TRANSCODER_TEMPLATE_ID ?= hls-adaptive-thumbnails-v2

help:
	@echo "Available commands:"
	@echo "  make check             - Build, vet, and test (CI / pre-commit gate)"
	@echo "  make test              - Run unit tests"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-verbose      - Run all tests with verbose output"
	@echo "  make setup             - Setup GCP infrastructure"
	@echo "  make verify            - Verify setup"
	@echo "  make run-api           - Run API server"
	@echo "  make run-processor     - Run processor service"
	@echo "  make clean             - Clean test cache"

check:
	@echo "Building..."
	@go build ./...
	@echo "Vetting..."
	@go vet ./...
	@echo "Testing..."
	@go test ./... -short

test:
	@echo "Running unit tests..."
	@go test ./internal/... -short -v

test-integration:
	@echo "Running integration tests..."
	@go test ./tests/integration/... -v

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./internal/... -short -coverpkg=./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-verbose:
	@echo "Running tests with verbose output..."
	@go test ./... -v -race -count=1

setup:
	@echo "Setting up GCP infrastructure..."
	@./scripts/setup-pubsub.sh
	@./scripts/setup-transcoder.sh
	@./scripts/setup-buckets.sh

verify:
	@echo "Verifying setup..."
	@gcloud services list --enabled | grep -q pubsub.googleapis.com && echo "✓ Pub/Sub API enabled" || echo "✗ Pub/Sub API not enabled"
	@gcloud services list --enabled | grep -q transcoder.googleapis.com && echo "✓ Transcoder API enabled" || echo "✗ Transcoder API not enabled"
	@gcloud pubsub topics describe video-uploaded >/dev/null 2>&1 && echo "✓ video-uploaded topic exists" || echo "✗ video-uploaded topic missing"
	@gcloud transcoder templates describe $(TRANSCODER_TEMPLATE_ID) --location=$(TRANSCODER_REGION) >/dev/null 2>&1 && echo "✓ Transcoder template exists" || echo "✗ Transcoder template missing"

run-api:
	@echo "Starting API server..."
	@go run cmd/api/main.go

run-processor:
	@echo "Starting processor service..."
	@go run cmd/processor/main.go

clean:
	@echo "Cleaning test cache..."
	@go clean -testcache
	@rm -f coverage.out coverage.html
