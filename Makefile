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

PROJECTNAME := $(shell basename "$(PWD)")
GOFILES := $(wildcard controller/*.go)
GOBIN := $(GOBASE)/bin


#CMDS=cosi-controller-manager
all:  unit build
#.PHONY: reltools
reltools: release-tools/build.make
release-tools/build.make:
	echo "TODO: update kubernetes/test-infra when controller and sidecar can build successfully"

build:
test:
unit:
codegen:
	@echo "Running update-codegen to generate the code..."
	bash ./hack/update-codegen.sh

	@echo "Running update-crd to generate the crd..."
	bash ./hack/update-crd.sh

include release-tools/build.make
