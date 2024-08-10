.DEFAULT_GOAL := help
SHELL = /usr/bin/env bash
THIS_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# 'go env' vars aren't always available in make environments, so get defaults for needed ones
GOARCH ?= $(shell go env GOARCH)

##
## ==== ARGS ===== #

## Container build tool compatible with `docker` API
DOCKER ?= docker

## Platform for 'build'
PLATFORM ?= linux/$(GOARCH)

## Image tag for all builds
IMAGE_TAG ?= cosi-provisioner-sidecar:latest

##
## === TARGETS === #

.PHONY: build
## Build local image for development, defaulting linux/<hostarch>
build:
	cd .. && $(DOCKER) build --file $(THIS_DIR)/Dockerfile --platform $(PLATFORM) --tag $(IMAGE_TAG) .

.PHONY: test
## Test packages
test:
	go vet ./...
	go test ./...

# print out lines beginning with double-comments, plus next line as basic help text
.PHONY: help
## Show this help text
help:
	@sed -n -e "/^##/{N;s/^/\n/p;}" $(MAKEFILE_LIST)
