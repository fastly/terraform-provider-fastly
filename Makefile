GO_BIN ?= go

PKG_NAME := fastly
VERSION := $(shell git describe --tags --always)
VERSION_SHORT := $(shell git describe --tags --always --abbrev=0 2>/dev/null || echo v0.0.0)
DOCS_PROVIDER_VERSION := $(subst v,,$(VERSION_SHORT))
BIN_DIR := $(CURDIR)/bin
BINARY := $(BIN_DIR)/terraform-provider-$(PKG_NAME)_$(VERSION)
OVERRIDES_FILE := $(BIN_DIR)/developer_overrides.tfrc

.PHONY: fmt build dev-overrides clean test-unit test-acc generate-docs validate-docs docs

fmt:
	@echo "==> Formatting Go code..."
	@$(GO_BIN) fmt ./...

build: fmt
	@echo "==> Building provider binary..."
	@mkdir -p $(BIN_DIR)
	@$(GO_BIN) build -o $(BINARY)
	@$(MAKE) --no-print-directory dev-overrides

dev-overrides:
	@echo "==> Generating Terraform CLI development overrides..."
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)

test-unit:
	@echo "==> Running unit tests..."
	@$(GO_BIN) test ./internal/...

test-acc:
	@echo "==> Running acceptance tests..."
	@echo "    Note: This requires FASTLY_API_TOKEN to be set"
	@TF_ACC=1 $(GO_BIN) test -v -timeout 30m ./internal/acceptance_tests -run TestAcc

generate-docs:
	@echo "==> Generating Terraform Registry documentation..."
	@$(GO_BIN) tool -modfile=tools/go.mod tfplugindocs generate --provider-name $(PKG_NAME)

validate-docs:
	@echo "==> Validating Terraform Registry documentation..."
	@$(GO_BIN) tool -modfile=tools/go.mod tfplugindocs validate --provider-name $(PKG_NAME)

docs: generate-docs validate-docs
