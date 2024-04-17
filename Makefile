.PHONY: all build clean default help init format check-license
default: help

BUILD ?= $(shell date +%FT%T%z)
GOVERSION ?= $(shell go version | cut -d " " -f3)
COMPONENT_VERSION ?= $(shell git describe --abbrev=0 --always)
COMPONENT_BRANCH ?= $(shell git describe --always --contains --all)
RELEASE_FULLCOMMIT ?= $(shell git rev-parse HEAD)
GO_BUILD_LDFLAGS = -X main.version=${COMPONENT_VERSION} -X main.buildDate=${BUILD} -X main.commit=${RELEASE_FULLCOMMIT} -X main.Branch=${COMPONENT_BRANCH} -X main.GoVersion=${GOVERSION} -s -w
NAME ?= mobserver

export RELEASE_PATH?=.

env:
	@echo $(TEST_ENV) | tr ' ' '\n' >.env

init:                       ## Install linters.
	cd tools && go generate -x -tags=tools

build:                      ## Compile using plain go build
	go build -ldflags="$(GO_BUILD_LDFLAGS)"  -o $(RELEASE_PATH)/$(NAME)

FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

format:                     ## Format source code.
	go mod tidy
	bin/gofumpt -l -w $(FILES)
	bin/gci write --section Standard --section Default --section "Prefix(github.com/kakao/$(NAME))" .

check:                      ## Run checks/linters
	bin/golangci-lint run

help:                       ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

test: env                   ## Run all tests.
	go test -v -count 1 -timeout 30s ./...

test-race: env              ## Run all tests with race flag.
	go test -race -v -timeout 30s ./...
