# Deploying Container Object Storage Interface (COSI) Provisioner Sidecar On Kubernetes

This document describes steps for Kubernetes administrators to setup Container Object Storage Interface (COSI) Provisioner Sidecar onto a Kubernetes cluster.

COSI Provisioner Sidecar can be setup using the [kustomization file](https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar/blob/master/kustomization.yaml) from the [container-object-storage-interface-provisioner-sidecar](https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar) repository with following command:

```sh
  kubectl create -k github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar
```
The output should look like the following:
```sh
namespace/objectstorage-provisioner-ns created
serviceaccount/objectstorage-provisioner-sa created
clusterrole.rbac.authorization.k8s.io/objectstorage-provisioner-role created
clusterrolebinding.rbac.authorization.k8s.io/objectstorage-provisioner-role-binding created
secret/objectstorage-provisioner created
deployment.apps/objectstorage-provisioner created
```

The Provisioner Sidecar will be deployed in the `default` namespace.

