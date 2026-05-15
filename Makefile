# ==============================================================================
# Configuration Variables
# ==============================================================================
BINARY_DIR    := bin
BINARY_SERVER := $(BINARY_DIR)/server
BINARY_CLIENT := $(BINARY_DIR)/client.exe

SRC_SERVER    := ./cmd/server
SRC_CLIENT    := ./cmd/client

# Optimization flags to trim debugging symbols and shrink binary footprint
LDFLAGS       := -ldflags="-s -w"

# ==============================================================================
# Production Automation Targets
# ==============================================================================
.PHONY: all build build-server build-client clean test help

all: clean build ## Clean workspace and compile all binaries natively

build: build-server build-client ## Compile both server (host target) and client (Windows target)

build-server: | $(BINARY_DIR) ## Build the Echo management control server natively for your Linux host
	@echo "==> Compiling control server for host OS..."
	@go build $(LDFLAGS) -o $(BINARY_SERVER) $(SRC_SERVER)
	@cp config.txt $(BINARY_DIR)/
	@echo "      [+] Server built successfully at: $(BINARY_SERVER)"

build-client: | $(BINARY_DIR) ## Cross-compile the client daemon executable specifically for the Windows VirtualBox environment
	@echo "==> Cross-compiling administration daemon for Windows (amd64)..."
	@env GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_CLIENT) $(SRC_CLIENT)
	@echo "      [+] Windows daemon compiled at: $(BINARY_CLIENT)"

$(BINARY_DIR):
	@mkdir -p $(BINARY_DIR)

clean: ## Scrape compiled executables out of your repository workspace
	@echo "==> Purging built artifacts..."
	@rm -f $(BINARY_SERVER) $(BINARY_CLIENT)
	@echo "      [+] Workspace clean."

test: ## Execute project-wide Go unit tests aggressively
	@echo "==> Triggering backend validation tests..."
	@go test -v -race ./...

help: ## Show this dynamic interactive target guide
	@echo "Available administrative commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'