# Makefile for pyqol

# Variables
BINARY_NAME := pyqol
GO_MODULE := github.com/pyqol/pyqol
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date +%Y-%m-%d)
LDFLAGS := -ldflags "-X '$(GO_MODULE)/internal/version.Version=$(VERSION)' \
                     -X '$(GO_MODULE)/internal/version.Commit=$(COMMIT)' \
                     -X '$(GO_MODULE)/internal/version.Date=$(DATE)' \
                     -X '$(GO_MODULE)/internal/version.BuiltBy=make'"

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

.PHONY: all build test clean install run version help build-python python-wheel python-test python-clean

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## all: Build and test
all: test build

## build: Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME) $(VERSION)...$(NC)"
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pyqol

## test: Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

## coverage: Generate coverage report
coverage:
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/
	go clean

## install: Install the binary
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	go install $(LDFLAGS) ./cmd/pyqol

## run: Run the application
run:
	go run $(LDFLAGS) ./cmd/pyqol

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	gofmt -s -w .

## lint: Run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	go vet ./...
	golangci-lint run

## release: Create a new release (use: make release VERSION=v0.1.0)
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(YELLOW)Please specify VERSION. Usage: make release VERSION=v0.1.0$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating release $(VERSION)...$(NC)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "$(GREEN)Release $(VERSION) created and pushed!$(NC)"

## dev: Development build with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Platform-specific builds
## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "$(GREEN)Building for Linux...$(NC)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/pyqol
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/pyqol

build-darwin:
	@echo "$(GREEN)Building for macOS...$(NC)"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/pyqol
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/pyqol

build-windows:
	@echo "$(GREEN)Building for Windows...$(NC)"
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/pyqol
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe ./cmd/pyqol

# Python packaging
## build-python: Build Python wheels with embedded binaries
build-python:
	@echo "$(GREEN)Building Python wheels...$(NC)"
	python/scripts/build_all_wheels.sh

## python-wheel: Build Python wheel for current platform only
python-wheel:
	@echo "$(GREEN)Building Python wheel for current platform...$(NC)"
	@mkdir -p python/src/pyqol/bin python/dist
	go build $(LDFLAGS) -ldflags="-s -w" -o python/src/pyqol/bin/pyqol-$$(go env GOOS)-$$(go env GOARCH)$$(if [ "$$(go env GOOS)" = "windows" ]; then echo ".exe"; fi) ./cmd/pyqol
	python/scripts/create_wheel.sh

## python-test: Test Python package installation
python-test: python-wheel
	@echo "$(GREEN)Testing Python package...$(NC)"
	cd python && pip install --force-reinstall dist/*.whl
	@echo "$(GREEN)Testing pyqol command...$(NC)"
	pyqol --version || pyqol --help

## python-clean: Clean Python build artifacts  
python-clean:
	@echo "$(YELLOW)Cleaning Python build artifacts...$(NC)"
	rm -rf python/src/pyqol/bin
	rm -rf python/dist
	rm -rf python/build
	rm -rf python/*.egg-info