.PHONY: build tidy clean version

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='-s -w' -o BUILD/nav-$(GOOS)-$(GOARCH) ./cmd/nav

tidy:
	go mod tidy

version:
	@./BUILD/nav-$(GOOS)-$(GOARCH) -version 2>&1 | awk '/^version:/{print $$2}'

clean:
	git clean -xfd
