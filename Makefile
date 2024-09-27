# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.DEFAULT_GOAL := help
.SUFFIXES: # remove legacy builtin suffixes to allow easier make debugging
SHELL = /usr/bin/env bash

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# If GOARCH is not set in the env, find it
GOARCH ?= $(shell go env GOARCH)

##
## ==== ARGS ===== #

## Container build tool compatible with `docker` API
DOCKER ?= docker

## Platform for 'build'
PLATFORM ?= linux/$(GOARCH)

## Additional args for 'build'
BUILD_ARGS ?=

## Image tag for controller image build
CONTROLLER_TAG ?= cosi-controller:latest

## Image tag for sidecar image build
SIDECAR_TAG ?= cosi-provisioner-sidecar:latest

##@ Development

.PHONY: all .gen
.gen: generate codegen # can be done in parallel with 'make -j'
.NOTPARALLEL: all # codegen must be finished before fmt/vet
all: .gen fmt vet build ## Build all targets, plus their prerequisites (faster with 'make -j')

.PHONY: generate
generate: controller/Dockerfile sidecar/Dockerfile ## Generate files
	$(MAKE) -C client crds
	$(MAKE) -C proto generate

.PHONY: codegen
codegen: codegen.client codegen.proto ## Generate code

.PHONY: fmt
fmt: fmt.client fmt.controller fmt.sidecar ## Format code

.PHONY: vet
vet: vet.client vet.controller vet.sidecar ## Vet code

.PHONY: test
test: .test.proto test.client test.controller test.sidecar ## Run tests including unit tests

.PHONY: test-e2e
test-e2e: # Run e2e tests
	@echo "unimplemented placeholder"

.PHONY: lint
lint: golangci-lint.client golangci-lint.controller golangci-lint.sidecar ## Run all linters (suggest `make -k`)

.PHONY: lint-fix
lint-fix: golangci-lint-fix.client golangci-lint-fix.controller golangci-lint-fix.sidecar ## Run all linters and perform fixes where possible (suggest `make -k`)


##@ Build

.PHONY: build
build: build.controller build.sidecar ## Build all container images for development

.PHONY: build.controller build.sidecar
build.controller: controller/Dockerfile ## Build only the controller container image
	$(DOCKER) build --file controller/Dockerfile --platform $(PLATFORM) $(BUILD_ARGS) --tag $(CONTROLLER_TAG) .
build.sidecar: sidecar/Dockerfile ## Build only the sidecar container image
	$(DOCKER) build --file sidecar/Dockerfile --platform $(PLATFORM) $(BUILD_ARGS) --tag $(SIDECAR_TAG) .

.PHONY: clean
## Clean build environment
clean:
	$(MAKE) -C proto clean

.PHONY: clobber
## Clean build environment and cached tools
clobber:
	$(MAKE) -C proto clobber
	rm -rf $(CURDIR)/.cache

##
## === TOOLS === #

GOLANGCI_LINT_VERSION ?= v1.61.0

TOOLBIN ?= $(CURDIR)/.cache/tools
$(TOOLBIN):
	mkdir -p $(TOOLBIN)

GOLANGCI_LINT ?= $(TOOLBIN)/golangci-lint
# .PHONY: golangci-lint
# golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(TOOLBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(TOOLBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

##
## === INTERMEDIATES === #

%/Dockerfile: hack/Dockerfile.in hack/gen-dockerfile.sh
	hack/gen-dockerfile.sh $* > "$@"

codegen.%: FORCE
	$(MAKE) -C $* codegen

fmt.%: FORCE
	cd $* && go fmt ./...

vet.%: FORCE
	cd $* && go vet ./...

test.%: fmt.% vet.% FORCE
	cd $* && go test ./...

# golangci-lint --new flag only complains about new code
golangci-lint.%: $(GOLANGCI_LINT)
	cd $* && $(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yaml --new

golangci-lint-fix.%: $(GOLANGCI_LINT)
	cd $* && $(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yaml --new --fix

.PHONY: .test.proto
.test.proto: # gRPC proto has a special unit test
	$(MAKE) -C proto check

.PHONY: FORCE # use this to force phony behavior for targets with pattern rules
FORCE:
