GO_BIN ?= go ## Allows overriding go executable.

TEST?=$$($(GO_BIN) list ./...)
GOFMT_FILES?=$$(find . -name '*.go')
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=fastly
FULL_PKG_NAME=github.com/fastly/terraform-provider-fastly
VERSION_PLACEHOLDER=version.ProviderVersion
VERSION=$(shell git describe --tags --always)
VERSION_SHORT=$(shell git describe --tags --always --abbrev=0)
DOCS_PROVIDER_VERSION=$(subst v,,$(VERSION_SHORT))

# Enables support for tools such as https://github.com/rakyll/gotest
TEST_COMMAND ?= $(GO_BIN) test

# R019: ignore large number of arguments passed to HasChanges().
# R018: replace sleep with either resource.Retry() or WaitForState().
# R001: for complex d.Set() calls use a string literal instead.
TFPROVIDERLINT_DEFAULT_FLAGS=-R001=false -R018=false -R019=false -XR001=false

# XAT001: missing resource.TestCase ErrorCheck.
TFPROVIDERLINTX_DEFAULT_FLAGS=-XAT001=false

GOHOSTOS ?= $(shell $(GO_BIN) env GOHOSTOS || echo unknown)
GOHOSTARCH ?= $(shell $(GO_BIN) env GOHOSTARCH || echo unknown)

# Use a parallelism of 4 by default for tests, overriding whatever GOMAXPROCS is
# set to. For the acceptance tests especially, the main bottleneck affecting the
# tests is network bandwidth and Fastly API rate limits. Therefore using the
# system default value of GOMAXPROCS, which is usually determined by the number
# of processors available, doesn't make the most sense.
TEST_PARALLELISM?=4

default: build

build: clean
	$(GO_BIN) build -o bin/terraform-provider-$(PKG_NAME)_$(VERSION) -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(VERSION)"
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

test:
	$(TEST_COMMAND) $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 $(TEST_COMMAND) $(TESTARGS) -timeout=30s -parallel=$(TEST_PARALLELISM)

# prefix `go test` with TF_LOG=debug or 'trace' for additional terraform output
# such as all the requests and responses it handles.
#
# reference:
# https://www.terraform.io/docs/internals/debugging.html
#
# TF_ACC is a recognised Terraform configuration that indicates we want to run an acceptance test, and which will result in creating REAL resources.
#
# reference:
# https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html#running-acceptance-tests
#
testacc: lint tfproviderlintx fmt
	TF_ACC=1 $(TEST_COMMAND) $(TEST) -v $(TESTARGS) -parallel=$(TEST_PARALLELISM) -timeout 360m -ldflags="-X=$(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=acc"

# WARNING: This target will delete infrastructure.
clean_test:
	@printf 'WARNING: This will delete infrastructure. Continue? (y/n) '; \
	read answer; \
	if echo "$$answer" | grep -iq '^y'; then \
	  SILENCE=true make sweep || true; \
		fastly service list --token $$FASTLY_API_KEY | grep -E '^tf\-' | awk '{print $$2}' | xargs -I % fastly service delete --token $$FASTLY_API_KEY -f -s %; \
		TEST_PARALLELISM=8 make testacc; \
	fi

fmt:
	golangci-lint fmt

goreleaser-bin:
	$(GO_BIN) get -modfile=tools.mod -tool github.com/goreleaser/goreleaser/v2@latest

# You can pass flags to goreleaser via GORELEASER_ARGS
# --skip=validate will skip the checks
# --clean will save you deleting the dist dir
# --single-target will be quicker and only build for your os & architecture
# e.g.
# make goreleaser GORELEASER_ARGS="--skip=validate --clean"
goreleaser: goreleaser-bin
	@GOHOSTOS="${GOHOSTOS}" GOHOSTARCH="${GOHOSTARCH}" $(GO_BIN) tool -modfile=tools.mod goreleaser build ${GORELEASER_ARGS}

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	$(TEST_COMMAND) -c $(TEST) $(TESTARGS)

generate-docs:
	$(shell sed -e "s/__VERSION__/$(DOCS_PROVIDER_VERSION)/g" examples/index-fastly-provider.tf.tmpl > examples/index-fastly-provider.tf)
	$(GO_BIN) tool -modfile=tools.mod tfplugindocs generate
	rm examples/index-fastly-provider.tf

validate-docs:
	$(GO_BIN) tool -modfile=tools.mod tfplugindocs validate

tfproviderlintx:
	$(GO_BIN) tool -modfile=tools.mod tfproviderlintx $(TFPROVIDERLINT_DEFAULT_FLAGS) $(TFPROVIDERLINTX_DEFAULT_FLAGS) $(TFPROVIDERLINTX_ARGS) ./...

tfproviderlint:
	$(GO_BIN) tool -modfile=tools.mod tfproviderlint $(TFPROVIDERLINT_DEFAULT_FLAGS) $(TFPROVIDERLINT_ARGS) ./...

sweep:
	@if [ "$(SILENCE)" != "true" ]; then \
		echo "WARNING: This will destroy infrastructure. Use only in development accounts."; \
	fi
	$(TEST_COMMAND) ./fastly -v -sweep=ALL $(SWEEPARGS) -timeout 30m || true

clean:
	rm -rf ./bin

validate-interface:
	@./tests/interface/script.sh

lint:
	golangci-lint run --verbose

.PHONY: all build clean clean_test default errcheck fmt fmtcheck generate-docs goreleaser goreleaser-bin lint sweep test test-compile testacc validate-docs validate-interface vet
