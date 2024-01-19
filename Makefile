# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./cmd
BINARY_NAME := leon-speedcams-go

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
.PHONY: build
build:
	go build -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## build/prod: build the application for production in Linux, Windows and MacOS with stripped binaries
.PHONY: build/prod
build/prod:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o=./bin/${BINARY_NAME}-linux-amd64 ${MAIN_PACKAGE_PATH}
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o=./bin/${BINARY_NAME}-windows-amd64.exe ${MAIN_PACKAGE_PATH}
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o=./bin/${BINARY_NAME}-darwin-amd64 ${MAIN_PACKAGE_PATH}

## run: run the  application
.PHONY: run
run: build
	/tmp/bin/${BINARY_NAME}

