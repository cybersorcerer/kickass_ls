# Makefile for c64.nvim LSP Server and Test Client
# Supports cross-compilation for multiple platforms

# Project info
PROJECT_NAME := c64.nvim
SERVER_NAME := kickass_ls
CLIENT_NAME := kickass_cl
CLIENT_BIN := kickass_cl
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build settings
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/build
GOFILES := $(wildcard *.go internal/**/*.go)
CLIENT_FILES := $(wildcard kickass_cl/*.go)

# Build flags
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"
BUILD_FLAGS := -trimpath

# Cross-compilation targets
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

# Color output
BOLD := \033[1m
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RESET := \033[0m

.PHONY: all help clean build build-server build-client install \
	build-all build-darwin build-linux build-windows \
	test test-integration test-client \
	release clean-bin version

# Default target
all: build

help: ## Show this help message
	@echo "$(BOLD)$(CYAN)c64.nvim Build System$(RESET)"
	@echo ""
	@echo "$(BOLD)Usage:$(RESET)"
	@echo "  make [target]"
	@echo ""
	@echo "$(BOLD)Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'

version: ## Display version information
	@echo "$(BOLD)Version:$(RESET) $(VERSION)"
	@echo "$(BOLD)Build Date:$(RESET) $(BUILD_DATE)"
	@echo "$(BOLD)Go Version:$(RESET) $(shell go version)"

clean: ## Clean all build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -rf build/
	@rm -rf dist/
	@echo "$(GREEN)✓ Clean complete$(RESET)"

clean-bin: ## Clean only binaries, keep source files
	@echo "$(YELLOW)Cleaning binaries...$(RESET)"
	@rm -rf build/
	@echo "$(GREEN)✓ Binaries cleaned$(RESET)"

# Local build targets (current platform)
build: build-server build-client ## Build both server and client for current platform

build-server: $(GOFILES) ## Build LSP server for current platform
	@echo "$(CYAN)Building LSP server ($(SERVER_NAME))...$(RESET)"
	@mkdir -p $(GOBIN)
	@go build $(BUILD_FLAGS) $(LDFLAGS) -o $(GOBIN)/$(SERVER_NAME) .
	@echo "$(GREEN)✓ Server built: $(GOBIN)/$(SERVER_NAME)$(RESET)"

build-client: $(CLIENT_FILES) ## Build test client for current platform
	@echo "$(CYAN)Building test client ($(CLIENT_NAME))...$(RESET)"
	@mkdir -p $(GOBIN)
	@cd kickass_cl && go build $(BUILD_FLAGS) $(LDFLAGS) -o $(GOBIN)/$(CLIENT_BIN) .
	@echo "$(GREEN)✓ Client built: $(GOBIN)/$(CLIENT_BIN)$(RESET)"

install: build ## Install server to ~/.local/bin and config to ~/.config/kickass_ls/
	@echo "$(CYAN)Installing $(SERVER_NAME)...$(RESET)"
	@mkdir -p ~/.local/bin
	@mkdir -p ~/.config/kickass_ls
	@cp $(GOBIN)/$(SERVER_NAME) ~/.local/bin/
	@cp kickass.json ~/.config/kickass_ls/
	@cp mnemonic.json ~/.config/kickass_ls/
	@cp c64memory.json ~/.config/kickass_ls/
	@echo "$(GREEN)✓ Binary installed to ~/.local/bin/$(SERVER_NAME)$(RESET)"
	@echo "$(GREEN)✓ Config installed to ~/.config/kickass_ls/$(RESET)"
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo "$(YELLOW)⚠ Warning: ~/.local/bin is not in your PATH$(RESET)"; \
		echo "$(YELLOW)  Add this line to your ~/.bashrc or ~/.zshrc:$(RESET)"; \
		echo "$(YELLOW)  export PATH=\"\$$HOME/.local/bin:\$$PATH\"$(RESET)"; \
	fi

# Cross-compilation targets
build-all: $(PLATFORMS) ## Build for all supported platforms
	@echo "$(GREEN)✓ All platforms built successfully$(RESET)"

darwin/amd64: ## Build for macOS Intel (x86_64)
	@echo "$(CYAN)Building for macOS Intel (amd64)...$(RESET)"
	@mkdir -p dist/darwin-amd64
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/darwin-amd64/$(SERVER_NAME) .
	@cd kickass_cl && GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o ../dist/darwin-amd64/$(CLIENT_NAME) .
	@echo "$(GREEN)✓ macOS Intel build complete$(RESET)"

darwin/arm64: ## Build for macOS Apple Silicon (ARM64)
	@echo "$(CYAN)Building for macOS Apple Silicon (arm64)...$(RESET)"
	@mkdir -p dist/darwin-arm64
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/darwin-arm64/$(SERVER_NAME) .
	@cd kickass_cl && GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o ../dist/darwin-arm64/$(CLIENT_NAME) .
	@echo "$(GREEN)✓ macOS ARM64 build complete$(RESET)"

linux/amd64: ## Build for Linux x86_64
	@echo "$(CYAN)Building for Linux amd64...$(RESET)"
	@mkdir -p dist/linux-amd64
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/linux-amd64/$(SERVER_NAME) .
	@cd kickass_cl && GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o ../dist/linux-amd64/$(CLIENT_NAME) .
	@echo "$(GREEN)✓ Linux amd64 build complete$(RESET)"

linux/arm64: ## Build for Linux ARM64 (Raspberry Pi 4+, etc.)
	@echo "$(CYAN)Building for Linux arm64...$(RESET)"
	@mkdir -p dist/linux-arm64
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/linux-arm64/$(SERVER_NAME) .
	@cd kickass_cl && GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o ../dist/linux-arm64/$(CLIENT_NAME) .
	@echo "$(GREEN)✓ Linux ARM64 build complete$(RESET)"

windows/amd64: ## Build for Windows x86_64
	@echo "$(CYAN)Building for Windows amd64...$(RESET)"
	@mkdir -p dist/windows-amd64
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/windows-amd64/$(SERVER_NAME).exe .
	@cd kickass_cl && GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o ../dist/windows-amd64/$(CLIENT_NAME).exe .
	@echo "$(GREEN)✓ Windows amd64 build complete$(RESET)"

# Convenience targets for building all binaries per OS
build-darwin: darwin/amd64 darwin/arm64 ## Build all macOS variants
	@echo "$(GREEN)✓ All macOS builds complete$(RESET)"

build-linux: linux/amd64 linux/arm64 ## Build all Linux variants
	@echo "$(GREEN)✓ All Linux builds complete$(RESET)"

build-windows: windows/amd64 ## Build all Windows variants
	@echo "$(GREEN)✓ All Windows builds complete$(RESET)"

# Release packaging
release: build-all ## Create release packages for all platforms
	@echo "$(CYAN)Creating release packages...$(RESET)"
	@mkdir -p dist/releases
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		dir="$(GOBASE)/dist/$$os-$$arch"; \
		pkg_dir="$(GOBASE)/dist/temp/$(SERVER_NAME)-$(VERSION)-$$os-$$arch"; \
		mkdir -p $$pkg_dir; \
		cp $$dir/$(SERVER_NAME)* $$pkg_dir/; \
		cp $(GOBASE)/kickass.json $(GOBASE)/mnemonic.json $(GOBASE)/c64memory.json $$pkg_dir/; \
		if [ "$$os" = "windows" ]; then \
			cp $(GOBASE)/install.ps1 $$pkg_dir/; \
			zipfile="$(GOBASE)/dist/releases/$(SERVER_NAME)-$(VERSION)-$$os-$$arch.zip"; \
			cd $$pkg_dir && zip -q -r $$zipfile . && cd $(GOBASE); \
			echo "$(GREEN)✓ Created $$zipfile$(RESET)"; \
		else \
			cp $(GOBASE)/install.sh $$pkg_dir/; \
			chmod +x $$pkg_dir/install.sh; \
			tarball="$(GOBASE)/dist/releases/$(SERVER_NAME)-$(VERSION)-$$os-$$arch.tar.gz"; \
			cd $$pkg_dir && tar -czf $$tarball . && cd $(GOBASE); \
			echo "$(GREEN)✓ Created $$tarball$(RESET)"; \
		fi; \
	done
	@rm -rf dist/temp
	@cd dist/releases && shasum -a 256 * > checksums.txt && cd ../..
	@echo "$(GREEN)✓ All release packages created in dist/releases/$(RESET)"
	@echo "$(GREEN)✓ Checksums generated in dist/releases/checksums.txt$(RESET)"

# Testing targets
test-integration: build ## Run integration tests with test client
	@echo "$(CYAN)Running integration tests...$(RESET)"
	@$(GOBIN)/$(CLIENT_BIN) -server $(GOBIN)/$(SERVER_NAME) -suite test-cases/test-files/test-encoding-suite.json
	@echo "$(GREEN)✓ Integration tests complete$(RESET)"

test-client: build-client ## Test client compilation only
	@echo "$(GREEN)✓ Test client built successfully$(RESET)"
	@$(GOBIN)/$(CLIENT_BIN) --help

# Development helpers
dev: clean build install ## Clean, build and install (full dev cycle)
	@echo "$(GREEN)✓ Development build complete$(RESET)"

fmt: ## Format Go code
	@echo "$(CYAN)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

vet: ## Run go vet
	@echo "$(CYAN)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet passed$(RESET)"

lint: fmt vet ## Run formatting and vetting

# Info targets
list-platforms: ## List all supported build platforms
	@echo "$(BOLD)Supported platforms:$(RESET)"
	@echo "$(PLATFORMS)" | tr ' ' '\n' | sed 's/^/  - /'

info: version list-platforms ## Show build information
	@echo ""
	@echo "$(BOLD)Project:$(RESET) $(PROJECT_NAME)"
	@echo "$(BOLD)Server:$(RESET) $(SERVER_NAME)"
	@echo "$(BOLD)Client:$(RESET) $(CLIENT_NAME)"
	@echo "$(BOLD)Go Root:$(RESET) $(GOBASE)"

# Check if required tools are installed
check-deps: ## Check if required build dependencies are installed
	@echo "$(CYAN)Checking dependencies...$(RESET)"
	@command -v go >/dev/null 2>&1 || { echo "$(YELLOW)⚠ Go is not installed$(RESET)"; exit 1; }
	@echo "$(GREEN)✓ Go found: $$(go version)$(RESET)"
	@command -v git >/dev/null 2>&1 || { echo "$(YELLOW)⚠ Git is not installed$(RESET)"; exit 1; }
	@echo "$(GREEN)✓ Git found: $$(git --version)$(RESET)"
	@echo "$(GREEN)✓ All dependencies satisfied$(RESET)"
