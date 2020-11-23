#!/bin/bash

SCRIPT_ROOT=$(dirname $0)

deepcopy-gen --input-dirs github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1 \
             --go-header-file hack/boilerplate.go.txt \
	     --output-base $GOPATH/src \
	     --output-file-base zz_generated.deepcopy \
	     --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1

openapi-gen --input-dirs github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1 \
            --go-header-file hack/boilerplate.go.txt \
	    --output-base $GOPATH/src \
	    --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1

defaulter-gen --input-dirs github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1 \
              --go-header-file hack/boilerplate.go.txt \
	      --output-base $GOPATH/src \
	      --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/defaulters

lister-gen --input-dirs github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1 \
              --go-header-file hack/boilerplate.go.txt \
	      --output-base $GOPATH/src \
	      --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/listers

informer-gen --input-dirs github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1 \
              --go-header-file hack/boilerplate.go.txt \
	      --output-base $GOPATH/src \
	      --listers-package github.com/kubernetes-sigs/container-object-storage-interface-api/listers \
	      --versioned-clientset-package github.com/kubernetes-sigs/container-object-storage-interface-api/clientset \
	      --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/informers

client-gen --input objectstorage.k8s.io/v1alpha1 \
           --go-header-file hack/boilerplate.go.txt \
	   --input-base github.com/kubernetes-sigs/container-object-storage-interface-api/apis/ \
	   --output-package github.com/kubernetes-sigs/container-object-storage-interface-api/ \
	   --output-base $GOPATH/src \
	   --clientset-name "clientset"

controller-gen crd:crdVersions=v1 paths=$SCRIPT_ROOT/../apis/... output:dir=$SCRIPT_ROOT/../crds
