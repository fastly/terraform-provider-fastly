TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=fastly
FULL_PKG_NAME=github.com/fastly/terraform-provider-fastly
VERSION_PLACEHOLDER=version.ProviderVersion
VERSION=$(shell git describe --tags --always)

# Use a parallelism of 4 by default for tests, overriding whatever GOMAXPROCS is
# set to. For the acceptance tests especially, the main bottleneck affecting the
# tests is network bandwidth and Fastly API rate limits. Therefore using the
# system default value of GOMAXPROCS, which is usually determined by the number
# of processors available, doesn't make the most sense.
TEST_PARALLELISM?=4

default: build

build:
	go build -o bin/terraform-provider-$(PKG_NAME)_$(VERSION) -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(VERSION)"
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=$(TEST_PARALLELISM)

# prefix `go test` with TF_LOG=debug or 'trace' for additional terraform output
# such as all the requests and responses it handles.
#
# reference:
# https://www.terraform.io/docs/internals/debugging.html
#
testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -parallel=$(TEST_PARALLELISM) -timeout 360m -ldflags="-X=$(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=acc"

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

BIN=$(CURDIR)/bin
$(BIN)/%:
	@echo "Installing tools from tools/tools.go"
	@cat tools/tools.go | grep _ | awk -F '"' '{print $$2}' | GOBIN=$(BIN) xargs -tI {} go install {}

generate-docs: $(BIN)/tfplugindocs
	go run -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(shell git describe --tags --always --abbrev=0)" scripts/generate-docs.go -tfplugindocsPath=$(BIN)/tfplugindocs

validate-docs: $(BIN)/tfplugindocs
	$(BIN)/tfplugindocs validate

tfproviderlintx: $(BIN)/tfproviderlintx
	$(BIN)/tfproviderlintx $(TFPROVIDERLINT_ARGS) ./...

tfproviderlint: $(BIN)/tfproviderlint
	$(BIN)/tfproviderlint $(TFPROVIDERLINT_ARGS) ./...

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./fastly -v -sweep=ALL $(SWEEPARGS) -timeout 30m

clean:
	rm -rf ./bin

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile sweep validate-docs generate-docs clean
