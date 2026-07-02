SHELL := /bin/sh

BINARY_NAME := gx
DIST_DIR := dist
GO ?= go
GOLANGCI_LINT ?= golangci-lint
VERSION ?= dev
GO_LDFLAGS := -X github.com/XDwanj/gx/internal/app.Version=$(VERSION)
GRAMMAR_TAGS := grammar_subset grammar_subset_bash grammar_subset_c grammar_subset_cpp grammar_subset_go grammar_subset_java grammar_subset_kotlin grammar_subset_lua grammar_subset_proto grammar_subset_python grammar_subset_ruby grammar_subset_rust grammar_subset_swift grammar_subset_typescript grammar_subset_tsx grammar_subset_zig

HOST_GOOS := $(shell $(GO) env GOOS)
HOST_GOARCH := $(shell $(GO) env GOARCH)
LOCAL_EXT :=

ifeq ($(HOST_GOOS),windows)
LOCAL_EXT := .exe
endif

LOCAL_BINARY := $(BINARY_NAME)$(LOCAL_EXT)
CROSS_TARGETS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: build test lint clean cross cross-darwin

build:
	$(GO) build -ldflags "$(GO_LDFLAGS)" -o $(LOCAL_BINARY) .

test:
	$(GO) test ./... -timeout 60s

lint:
	$(GOLANGCI_LINT) run ./...

clean:
	rm -rf $(LOCAL_BINARY) $(DIST_DIR)

cross:
	@mkdir -p $(DIST_DIR)
	@set -eu; \
	for target in $(CROSS_TARGETS); do \
		goos="$${target%/*}"; \
		goarch="$${target#*/}"; \
		ext=""; \
		case "$$goos/$$goarch" in \
			darwin/amd64) ;; \
			darwin/arm64) ;; \
			linux/amd64) ;; \
			linux/arm64) ;; \
			windows/amd64) ext='.exe' ;; \
			*) echo "unsupported cross target: $$target" >&2; exit 1 ;; \
		esac; \
		output="$(DIST_DIR)/$(BINARY_NAME)-$$goos-$$goarch$$ext"; \
		echo "building $$output"; \
		CGO_ENABLED=0 GOOS="$$goos" GOARCH="$$goarch" $(GO) build -tags "$(GRAMMAR_TAGS)" -ldflags "$(GO_LDFLAGS)" -o "$$output" .; \
	done

cross-darwin:
	$(MAKE) cross CROSS_TARGETS="darwin/amd64 darwin/arm64"
