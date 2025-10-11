# Makefile for pyscn

# Variables
BINARY_NAME := pyscn
GO_MODULE := github.com/ludo-technologies/pyscn
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date +%Y-%m-%d)
LDFLAGS := -ldflags "-s -w \
                     -X '$(GO_MODULE)/internal/version.Version=$(VERSION)' \
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
	@printf "$(GREEN)Building $(BINARY_NAME) $(VERSION)...$(NC)\n"
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pyscn

## test: Run tests
test:
	@printf "$(GREEN)Running tests...$(NC)\n"
	go test -v ./...

## bench: Run benchmarks
bench:
	@printf "$(GREEN)Running benchmarks...$(NC)\n"
	go test -bench=. -benchmem ./...

## coverage: Generate coverage report
coverage:
	@printf "$(GREEN)Generating coverage report...$(NC)\n"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)Coverage report generated: coverage.html$(NC)\n"

## clean: Clean build artifacts
clean:
	@printf "$(YELLOW)Cleaning...$(NC)\n"
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/
	go clean

## install: Install the binary
install: build
	@printf "$(GREEN)Installing $(BINARY_NAME)...$(NC)\n"
	go install $(LDFLAGS) ./cmd/pyscn

## run: Run the application
run:
	go run $(LDFLAGS) ./cmd/pyscn

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

## fmt: Format code
fmt:
	@printf "$(GREEN)Formatting code...$(NC)\n"
	go fmt ./...
	gofmt -s -w .

## lint: Run linters
lint:
	@printf "$(GREEN)Running linters...$(NC)\n"
	go vet ./...
	golangci-lint run

## release: Create a new release (use: make release VERSION=v0.1.0)
release:
	@if [ -z "$(VERSION)" ]; then \
		printf "$(YELLOW)Please specify VERSION. Usage: make release VERSION=v0.1.0$(NC)\n"; \
		exit 1; \
	fi
	@printf "$(GREEN)Creating release $(VERSION)...$(NC)\n"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@printf "$(GREEN)Release $(VERSION) created and pushed!$(NC)\n"

## dev: Development build with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Platform-specific builds
## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@printf "$(GREEN)Building for Linux...$(NC)\n"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/pyscn
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/pyscn

build-darwin:
	@printf "$(GREEN)Building for macOS...$(NC)\n"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/pyscn
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/pyscn

build-windows:
	@printf "$(GREEN)Building for Windows...$(NC)\n"
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/pyscn
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe ./cmd/pyscn

# Python packaging
## build-python: Build Python wheels with embedded binaries
build-python:
	@printf "$(GREEN)Building Python wheels...$(NC)\n"
	python/scripts/build_all_wheels.sh

## python-wheel: Build Python wheel for current platform only
python-wheel:
	@printf "$(GREEN)Building Python wheel for current platform...$(NC)\n"
	@mkdir -p python/src/pyscn/bin dist
	go build $(LDFLAGS) -o python/src/pyscn/bin/pyscn-$$(go env GOOS)-$$(go env GOARCH)$$(if [ "$$(go env GOOS)" = "windows" ]; then echo ".exe"; fi) ./cmd/pyscn
	python/scripts/create_wheel.sh

## python-test: Test Python package installation
python-test: python-wheel
	@printf "$(GREEN)Testing Python package...$(NC)\n"
	pip install --force-reinstall dist/*.whl
	@printf "$(GREEN)Testing pyscn command...$(NC)\n"
	pyscn --version || pyscn --help

## python-clean: Clean Python build artifacts
python-clean:
	@printf "$(YELLOW)Cleaning Python build artifacts...$(NC)\n"
	rm -rf python/src/pyscn/bin
	rm -rf dist
	rm -rf build
	rm -rf *.egg-info
