export

GOLANGCI_LINT_VERSION=v1.22.2

GIT_HASH=$(shell git rev-parse --short HEAD)

GOBIN:=${PWD}/bin
PATH:=${GOBIN}:${PATH}

REVIEWDOG_ARGS ?= -diff="git diff master" -tee

dependencies:
	go mod download
	go mod tidy

build:
	go build \
	  -ldflags "-X github.com/hatappi/gomodoro/cmd.commit=${GIT_HASH}" \
	  -o ./dist/gomodoro

.PHONY: tools
tools:
	@go generate -tags tools tools/tools.go

.PHONY: lint
lint:
	@${GOBIN}/golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	@${GOBIN}/golangci-lint run --fix ./...

test:
	@go test ./...

reviewdog:
	${GOBIN}/reviewdog ${REVIEWDOG_ARGS}
