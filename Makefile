PKG_NAME := fastly
VERSION := $(shell git describe --tags --always)
BIN_DIR := $(CURDIR)/bin
BINARY := $(BIN_DIR)/terraform-provider-$(PKG_NAME)_$(VERSION)
OVERRIDES_FILE := $(BIN_DIR)/developer_overrides.tfrc

.PHONY: fmt build dev-overrides clean

fmt:
	@echo "==> Formatting Go code..."
	@go fmt ./...

build: fmt
	@echo "==> Building provider binary..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BINARY)
	@$(MAKE) --no-print-directory dev-overrides

dev-overrides:
	@echo "==> Generating Terraform CLI development overrides..."
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
