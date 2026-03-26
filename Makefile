MODULE   := github.com/jsuykerbuyk/resumectl
BINARY   := resumectl
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION) -X $(MODULE)/internal/cli.Commit=$(COMMIT) -X $(MODULE)/internal/cli.BuildDate=$(DATE)"

COVERAGE_THRESHOLD := 80

# Install prefix: ~/.local for non-root, /usr/local for root.
# Override with PREFIX= on the command line.
ifeq ($(shell id -u),0)
  PREFIX  ?= /usr/local
else
  PREFIX  ?= $(HOME)/.local
endif
DESTDIR  ?=
BINDIR   ?= $(PREFIX)/bin
MANDIR   ?= $(PREFIX)/share/man

.DEFAULT_GOAL := help

##@ General
.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build
.PHONY: build
build: ## Build the binary
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/resumectl

.PHONY: install
install: build man ## Install binary and man pages to PREFIX (default: ~/.local or /usr/local)
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -d $(DESTDIR)$(MANDIR)/man1
	install -m 644 man/*.1 $(DESTDIR)$(MANDIR)/man1/
	@echo "Installed $(BINARY) to $(DESTDIR)$(BINDIR)/$(BINARY)"
	@echo "Installed man pages to $(DESTDIR)$(MANDIR)/man1/"

.PHONY: uninstall
uninstall: ## Remove installed binary and man pages
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(MANDIR)/man1/resumectl*.1
	@echo "Uninstalled $(BINARY) from $(DESTDIR)$(BINDIR)"

##@ Test
.PHONY: test
test: ## Run unit tests
	go test -race ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage (80% gate)
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@TOTAL=$$(go tool cover -func=coverage.out | grep ^total: | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $${TOTAL}%"; \
	if [ "$$(echo "$${TOTAL} < $(COVERAGE_THRESHOLD)" | bc -l)" -eq 1 ]; then \
		echo "FAIL: coverage $${TOTAL}% is below $(COVERAGE_THRESHOLD)% threshold"; \
		exit 1; \
	fi; \
	echo "PASS: coverage $${TOTAL}% meets $(COVERAGE_THRESHOLD)% threshold"

.PHONY: test-cover-html
test-cover-html: test-cover ## Open coverage report in browser
	go tool cover -html=coverage.out

.PHONY: integration
integration: ## Run integration tests
	go test -race -tags integration ./...

##@ Quality
.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: check
check: vet lint test-cover ## Run all checks (vet, lint, test with coverage)

##@ Documentation
.PHONY: man
man: build ## Generate man pages to man/
	./bin/$(BINARY) doc --man-dir man

.PHONY: docs-md
docs-md: build ## Generate markdown command reference to docs/cli/
	./bin/$(BINARY) doc --md-dir docs/cli

.PHONY: docs
docs: man docs-md ## Generate all documentation

##@ Workflow
.PHONY: smoke
smoke: build ## Smoke test against a real LLM provider (needs API key)
	@echo "=== Smoke Test ==="
	@SMOKE_DIR=$$(mktemp -d) && \
	echo "Vault: $$SMOKE_DIR" && \
	./bin/$(BINARY) vault init "$$SMOKE_DIR" && \
	echo "Acme Corp is hiring a Staff Software Engineer for our infrastructure team. Requirements: Go, Kubernetes, distributed systems. Preferred: Rust, NVMe. Remote. 200-260k USD + equity." > "$$SMOKE_DIR/posting.txt" && \
	./bin/$(BINARY) --vault "$$SMOKE_DIR" --provider grok parse-job "$$SMOKE_DIR/posting.txt" && \
	echo "--- Vault contents after parse-job ---" && \
	./bin/$(BINARY) --vault "$$SMOKE_DIR" vault ls && \
	./bin/$(BINARY) --vault "$$SMOKE_DIR" vault stats && \
	rm -rf "$$SMOKE_DIR" && \
	echo "=== Smoke Test PASSED ==="

##@ Clean
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out
