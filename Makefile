GOLANGCI_VERSION = 1.53.3
OUTDIR := bin
PATH := bin:$(PATH)
SHELL := env PATH=$(PATH) /bin/bash

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint

bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p $(OUTDIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v${GOLANGCI_VERSION}
	@mv bin/golangci-lint "$@"

## audit: tidy and vendor dependencies and format, vet, lint and test all code
.PHONY: audit
audit: fmt tidy lint vet test
	@echo 'Auditing code done'

## vet: vetting code
.PHONY: vet
vet:
	@echo 'Vetting code...'
	@go vet $(shell go list ./... | grep -v /vendor/|xargs echo)

## test: test all code
.PHONY: test
test:
	@echo 'Running tests...'
	@CGO_ENABLED=0 go test $(shell go list ./... | grep -v /vendor/|xargs echo) -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

## fmt: formatting code
fmt:
	@echo 'Formatting code...'
	@go fmt $(shell go list ./... | grep -v /vendor/|xargs echo)

## vendor: tidy dependencies
.PHONY: tidy
tidy:
	@echo 'Tidying and verifying module dependencies...'
	@go mod tidy
	@go mod verify

## lint: linting code
.PHONY: lint
lint: bin/golangci-lint ## Run linter
	@echo 'Linting code...'
	golangci-lint run
