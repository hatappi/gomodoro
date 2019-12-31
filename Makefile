GOLANGCI_LINT_VERSION=v1.22.2
LINT_BIN_PATH:=$(shell go env GOPATH)

GIT_HASH=$(shell git rev-parse --short HEAD)

dependencies:
	go mod download
	go mod tidy

build:
	go build \
	  -ldflags "-X github.com/hatappi/gomodoro/cmd.commit=${GIT_HASH}" \
	  -o ./dist/gomodoro

install-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${LINT_BIN_PATH} ${GOLANGCI_LINT_VERSION}

lint:
	${LINT_BIN_PATH}/golangci-lint run ./...
