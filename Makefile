SHELL := /bin/bash

BIN := iam-proxy
GOBIN := $(or $(GOBIN),$(shell pwd)/bin)

# Code coverage
COVERAGE_REPORT := coverage.out

# Dependency versions
GOMOCK_VERSION ?= v1.5.0
SWAGGER_VERSION ?= v0.27.0
GOWRAP_VERSION ?= v1.2.0
GO_IMPORT_LINT_VERSION ?= latest

VERSION ?= $(shell git describe --tags --always)
SHA ?= $(shell git rev-parse HEAD)
TAG := $(SHA)

help: # @HELP prints this message
help:
	@echo "VARIABLES:"
	@echo "  BIN = $(BIN)"
	@echo "  COVERAGE_REPORT = $(COVERAGE_REPORT)"
	@echo
	@echo "Versions:"
	@echo "  GOMOCK_VERSION         = $(GOMOCK_VERSION)"
	@echo "  SWAGGER_VERSION        = $(SWAGGER_VERSION)"
	@echo "  GOWRAP_VERSION         = $(GOWRAP_VERSION)"
	@echo "  GO_IMPORT_LINT_VERSION = $(GO_IMPORT_LINT_VERSION)"
	@echo
	@echo "TARGETS:"
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '

.PHONY: all
all: # @HELP builds the binary
all: build

.PHONY: build
build: # @HELP creates auto-generated files and builds binary
build: generate compile

.PHONY: compile
compile: # @HELP runs the actual `go build` which updates service binary.
compile:
	@echo "Building $(GOBIN)"
	VERSION=$(VERSION) GOBIN=$(GOBIN) GOWRAP_VERSION=$(GOWRAP_VERSION) ./build/build.sh

.PHONY: install-deps
install-deps: # @HELP downloads dependencies, as defined in ./build/install-deps.sh
install-deps: bin/mockgen bin/swagger bin/gowrap

bin/mockgen bin/swagger bin/gowrap:
	GOBIN=$(GOBIN) GOMOCK_VERSION=$(GOMOCK_VERSION) SWAGGER_VERSION=$(SWAGGER_VERSION) GOWRAP_VERSION=$(GOWRAP_VERSION) ./build/install-deps.sh

.PHONY: generate
generate: # @HELP downloads dependencies and runs go generate
generate: install-deps
	PATH=$(PATH):$(GOBIN) go generate -x ./...
	PATH=$(PATH):$(GOBIN) swagger validate docs/swagger/iam.json

.PHONY: test
test: # @HELP runs all tests and generates a RAW coverage report to be picked up by analysis tools
test:
	@echo -n "Running full tests... "
	@GOBIN=$(GOBIN) go test -race -cover -coverprofile=$(COVERAGE_REPORT) -covermode=atomic ./...
	@GOBIN=$(GOBIN) go tool cover -func=$(COVERAGE_REPORT) | tail -n 1
	@echo "done."

.PHONY: coverage-html
coverage-html: # @HELP generates an HTML coverage report
coverage-html: test
	@GOBIN=$(GOBIN) go tool cover -html=$(COVERAGE_REPORT)

version: # @HELP outputs the version string
version:
	@echo $(VERSION)

.PHONY: pre-commit
pre-commit: # @HELP runs pre-commit checks on the entire repo
pre-commit:
	@pre-commit run --all

.PHONY: go-import-lint
go-import-lint: # @HELP verifies the imports order
go-import-lint: bin/go-import-lint check-go-import-lint
	@echo "Verifying import order"
	@$(GOBIN)/go-import-lint

.PHONY: bin/go-import-lint
bin/go-import-lint:
	@echo "Downloading go-import-lint..."
	@GOBIN=$(GOBIN) go install -v github.com/hedhyw/go-import-lint/cmd/go-import-lint@$(GO_IMPORT_LINT_VERSION)

.PHONY: check-go-import-lint
check-go-import-lint:
	@which ${GOBIN}/go-import-lint >/dev/null || ( echo "Install go-import-lint from https://github.com/hedhyw/go-import-lint and retry." && exit 1 )

clean: # @HELP removes built binaries and temporary files
clean:
	rm -rf .go bin
	rm -rf docs/swagger/iam.json

build-image: # @HELP builds service image
build-image:
	docker build -t $(BIN) .

.PHONY: run
run: # @HELP run the built binary
run: build
	./bin/iam-proxy
