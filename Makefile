SHELL := /bin/sh

BINARY_NAME := gx
DIST_DIR := dist
GO ?= go
GOLANGCI_LINT ?= golangci-lint
VERSION ?= dev
GO_LDFLAGS := -X github.com/XDwanj/gx/internal/app.Version=$(VERSION)

HOST_GOOS := $(shell $(GO) env GOOS)
HOST_GOARCH := $(shell $(GO) env GOARCH)
LOCAL_EXT :=

ifeq ($(HOST_GOOS),windows)
LOCAL_EXT := .exe
endif

LOCAL_BINARY := $(BINARY_NAME)$(LOCAL_EXT)
CROSS_TARGETS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

CC_DARWIN_AMD64 ?= clang -arch x86_64
CC_DARWIN_ARM64 ?= clang -arch arm64
CC_LINUX_AMD64 ?= zig cc -target x86_64-linux-gnu
CC_LINUX_ARM64 ?= zig cc -target aarch64-linux-gnu
CC_WINDOWS_AMD64 ?= zig cc -target x86_64-windows-gnu

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
			darwin/amd64) cc='$(CC_DARWIN_AMD64)' ;; \
			darwin/arm64) cc='$(CC_DARWIN_ARM64)' ;; \
			linux/amd64) cc='$(CC_LINUX_AMD64)' ;; \
			linux/arm64) cc='$(CC_LINUX_ARM64)' ;; \
			windows/amd64) cc='$(CC_WINDOWS_AMD64)'; ext='.exe' ;; \
			*) echo "unsupported cross target: $$target" >&2; exit 1 ;; \
		esac; \
		cc_bin="$${cc%% *}"; \
		if ! command -v "$$cc_bin" >/dev/null 2>&1; then \
			echo "missing C compiler '$$cc_bin' for $$target; override the matching CC_* variable or install the toolchain" >&2; \
			exit 1; \
		fi; \
		output="$(DIST_DIR)/$(BINARY_NAME)-$$goos-$$goarch$$ext"; \
		echo "building $$output"; \
		CGO_ENABLED=1 CC="$$cc" GOOS="$$goos" GOARCH="$$goarch" $(GO) build -ldflags "$(GO_LDFLAGS)" -o "$$output" .; \
	done

cross-darwin:
	$(MAKE) cross CROSS_TARGETS="darwin/amd64 darwin/arm64"
