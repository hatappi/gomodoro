GOLANGCI_LINT_VERSION=v1.22.2
LINT_BIN_PATH:=$(shell go env GOPATH)

GIT_HASH=$(shell git rev-parse --short HEAD)

GOBIN:=${PWD}/bin
PATH:=${GOBIN}:${PATH}

dependencies:
	go mod download
	go mod tidy

build:
	go build \
	  -ldflags "-X github.com/hatappi/gomodoro/cmd.commit=${GIT_HASH}" \
	  -o ./dist/gomodoro

install-tools:
	@GOBIN=${GOBIN} ./scripts/install_tools.sh

lint:
	@${LINT_BIN_PATH}/golangci-lint run ./...

test:
	@go test ./...
