.PHONY: build lint tidy clean version hooks

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='-s -w' -o BUILD/nav-$(GOOS)-$(GOARCH) ./cmd/nav

lint:
	golangci-lint run --fix

tidy:
	go mod tidy

version:
	@./BUILD/nav-$(GOOS)-$(GOARCH) -v 2>&1 | awk '/^version:/{print $$2}'

hooks:
	git config core.hooksPath .githooks

clean:
	git clean -xfd
