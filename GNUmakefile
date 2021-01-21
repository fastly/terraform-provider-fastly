TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=fastly
FULL_PKG_NAME=github.com/fastly/terraform-provider-fastly
VERSION_PLACEHOLDER=version.ProviderVersion
VERSION=$(shell git describe --tags --always)

default: build

build:
	go install -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(VERSION)"

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

# prefix `go test` with TF_LOG=debug or 'trace' for additional terraform output
# such as all the requests and responses it handles.
#
# reference:
# https://www.terraform.io/docs/internals/debugging.html
#
testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 360m -ldflags="-X=$(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=acc"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"


test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
	go run "scripts/website/parse-templates.go"

website-test:
	go run "scripts/website/parse-templates.go"

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./fastly -v -sweep=ALL $(SWEEPARGS) -timeout 30m

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile website website-test sweep build_local clean install_local

PROVIDER_HOSTNAME=registry.terraform.io
PROVIDER_NAMESPACE=fastly
PROVIDER_TYPE=fastly
PROVIDER_TARGET=$(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_VERSION = 99.99.99

PLUGINS_PATH = ~/.terraform.d/plugins
PLUGINS_PROVIDER_PATH=$(PROVIDER_HOSTNAME)/$(PROVIDER_NAMESPACE)/$(PROVIDER_TYPE)/$(PROVIDER_VERSION)/$(PROVIDER_TARGET)

build_local:
	@echo "Building local provider binary"
	@mkdir -p ./bin
	go build -o bin/terraform-provider-$(PROVIDER_TYPE)_v$(PROVIDER_VERSION)

clean:
	@echo "Deleting local provider binary"
	rm -rf ./bin

install_local: build_local
	@echo "Installing local provider binary to plugins mirror path $(PLUGINS_PATH)/$(PLUGINS_PROVIDER_PATH)"
	@mkdir -p $(PLUGINS_PATH)/$(PLUGINS_PROVIDER_PATH)
	@cp ./bin/terraform-provider-$(PROVIDER_TYPE)_v$(PROVIDER_VERSION) $(PLUGINS_PATH)/$(PLUGINS_PROVIDER_PATH)
