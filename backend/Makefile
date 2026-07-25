# Pinned tool versions (run via `go run` so no global installs are needed).
GOLANGCI_VERSION := v2.5.0
GOVULNCHECK_VERSION := v1.1.4

GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
GOVULNCHECK := go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

APP_VERSION ?= 0.0.0

.PHONY: help fmt lint lint-fix mod-check test test-unit test-race vuln build docker-build check clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

fmt: ## Apply gofmt + goimports formatting
	$(GOLANGCI) fmt

lint: ## Run golangci-lint (errcheck, govet, staticcheck, unused, gosec + format check)
	$(GOLANGCI) run

lint-fix: ## Run golangci-lint with --fix
	$(GOLANGCI) run --fix

mod-check: ## Verify module integrity and that go.mod/go.sum are tidy
	go mod verify
	go mod tidy
	git diff --exit-code go.mod go.sum

test: test-unit test-race ## Run unit + race tests

test-unit: ## Run unit tests
	CGO_ENABLED=0 go test ./...

test-race: ## Run tests with the race detector
	CGO_ENABLED=1 go test -race ./...

vuln: ## Run govulncheck call-graph vulnerability scan
	$(GOVULNCHECK) ./...

build: ## Hardened static build
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.Version=$(APP_VERSION)" -o ./tmp/server ./cmd/server

docker-build: ## Build the production container image
	docker build -t my-hours-backend .

check: lint mod-check vuln test build ## Run the full pre-PR suite

clean: ## Remove build artifacts
	rm -rf ./tmp
