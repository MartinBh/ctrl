APP_NAME := ctrl
CMD := ./cmd/ctrl
DIST_DIR := dist
BIN := $(DIST_DIR)/$(APP_NAME)
CACHE_DIR := .cache
GO ?= go
GOCACHE ?= $(CURDIR)/$(CACHE_DIR)/go-build

export GOCACHE

.DEFAULT_GOAL := help

.PHONY: help run build install test vet fmt tidy check clean

help: ## Show available targets.
	@printf "Available targets:\n"
	@printf "  make run      Run the TUI locally\n"
	@printf "  make build    Build $(APP_NAME) into $(BIN)\n"
	@printf "  make install  Install $(APP_NAME) with go install\n"
	@printf "  make test     Run Go tests\n"
	@printf "  make vet      Run go vet\n"
	@printf "  make fmt      Format Go source files\n"
	@printf "  make tidy     Tidy Go module dependencies\n"
	@printf "  make check    Run fmt, tidy, test, and vet\n"
	@printf "  make clean    Remove local build artifacts\n"

run: ## Run the TUI locally.
	$(GO) run $(CMD)

build: ## Build the CLI into dist/.
	mkdir -p $(DIST_DIR)
	$(GO) build -o $(BIN) $(CMD)

install: ## Install the CLI with go install.
	$(GO) install $(CMD)

test: ## Run Go tests.
	$(GO) test ./...

vet: ## Run go vet.
	$(GO) vet ./...

fmt: ## Format Go source files.
	gofmt -w cmd internal

tidy: ## Tidy Go module dependencies.
	$(GO) mod tidy

check: fmt tidy test vet ## Run the standard verification chain.

clean: ## Remove local build artifacts.
	rm -rf $(DIST_DIR) $(CACHE_DIR) ./$(APP_NAME) ./cmd/ctrl/$(APP_NAME)
