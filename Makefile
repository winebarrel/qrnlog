SHELL   := /bin/bash
VERSION := v2.2.2
GOOS      := $(shell go env GOOS)
GOARCH    := $(shell go env GOARCH)

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/qrnlog

.PHONY: package
package: clean build
	gzip qrnlog -c > qrnlog_$(VERSION)_$(GOOS)_$(GOARCH).gz
	sha1sum qrnlog_$(VERSION)_$(GOOS)_$(GOARCH).gz > qrnlog_$(VERSION)_$(GOOS)_$(GOARCH).gz.sha1sum

.PHONY: clean
clean:
	rm -f qrnlog
