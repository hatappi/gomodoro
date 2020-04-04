GOLANGCI_LINT_VERSION=v1.22.2

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
	@${GOBIN}/golangci-lint run ./...

test:
	@go test ./...
