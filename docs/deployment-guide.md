---
title: Deploying Container Object Storage Interface (COSI) On Kubernetes
---
# Deploying Container Object Storage Interface (COSI) On Kubernetes

This document describes steps for Kubernetes administrators to setup Container Object Storage Interface (COSI) onto a Kubernetes cluster.
## Overview

Following components that need to be deployed in Kubernetes to setup COSI.

- CustomResourceDefinitions (CRDs)
- Controller
- Driver
- Sidecar for the driver
- Node Adapter

### Quick Start

Execute following commands to setup COSI:

```sh
# Install CRDs
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-api

# Install controller
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-controller

# Sample Provisioner and Sidecar
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar

# Node Adapter
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-csi-adapter
```

### CustomResourceDefinitions

COSI acts on following custom resource definitions (CRDs):

- `BucketRequest` - Represents a request to provision a Bucket
- `BucketClass` - Represents a class of Buckets with similar characteristics
- `Bucket` - Represents a Bucket or its equivalent in the storage backend
- `BucketAccessRequest` - Represents a request to access a Bucket
- `BucketAccessClass` - Represents a class of accessors with similar access requirements
- `BucketAccess` - Represents a access token or service account in the storage backend

All [COSI custom resource definitions](../crds) can be installed using [kustomization file](../kustomization.yaml) and `kubectl` with following command:

```sh
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-api
```

### Controller

COSI controller can be setup using the [kustomization file](https://github.com/kubernetes-sigs/container-object-storage-interface-controller/blob/master/kustomization.yaml) from the [container-object-storage-interface-controller](https://github.com/kubernetes-sigs/container-object-storage-interface-controller) repository with following command:

```sh
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-controller
```

The controller will be deployed in the `default` namespace.

### Sample Driver & Sidecar

Sample Driver & Sidecar can be setup using the [kustomization file](https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar/blob/master/kustomization.yaml) from the [container-object-storage-interface-provisioner-sidecar](https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar) repository with following command:

```sh
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar
```
### Node Adapter

Node adapter can be setup using the [kustomization file](https://github.com/kubernetes-sigs/container-object-storage-interface-csi-adapter/blob/master/kustomization.yaml) from the [container-object-storage-interface-csi-adapter](https://github.com/kubernetes-sigs/container-object-storage-interface-csi-adapter) repository with following command:

```sh
kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-csi-adapter
```
