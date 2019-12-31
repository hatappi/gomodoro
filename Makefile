GOLANGCI_LINT_VERSION=v1.22.2
LINT_BIN_PATH:=$(shell go env GOPATH)

install-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${LINT_BIN_PATH} ${GOLANGCI_LINT_VERSION}

lint:
	${LINT_BIN_PATH}/golangci-lint run ./...
