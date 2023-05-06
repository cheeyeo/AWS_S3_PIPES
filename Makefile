MAIN_PACKAGE_PATH := ./cmd
BINARY_NAME := s3pipe

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## build: build the application
.PHONY: build
build:
	go build -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## test: run all tests
.PHONY: test
test:
	go test -v ./tests/...
