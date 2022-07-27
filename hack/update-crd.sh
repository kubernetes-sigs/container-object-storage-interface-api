#!/bin/bash

SCRIPT_ROOT_RELATIVE=$(dirname "${BASH_SOURCE}")/..
SCRIPT_ROOT=$(realpath "${SCRIPT_ROOT_RELATIVE}")
CONTROLLERTOOLS_PKG=${CONTROLLERTOOLS_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/sigs.k8s.io/controller-tools 2>/dev/null || echo ../code-controller-tools)}

# find or download controller-gen
pushd "${CONTROLLERTOOLS_PKG}"
trap popd exit
go run -v ./cmd/controller-gen crd:crdVersions=v1 paths="${SCRIPT_ROOT}/apis/objectstorage/..." output:crd:dir="${SCRIPT_ROOT}/crds"