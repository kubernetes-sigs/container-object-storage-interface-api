#!/bin/bash

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

bash "${CODEGEN_PKG}"/generate-internal-groups.sh "deepcopy,client,informer,lister,openapi" \
  sigs.k8s.io/container-object-storage-interface-api/client \
  sigs.k8s.io/container-object-storage-interface-api/apis \
  sigs.k8s.io/container-object-storage-interface-api/apis \
  objectstorage:v1alpha1 \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"
