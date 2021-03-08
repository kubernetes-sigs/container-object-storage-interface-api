![version](https://img.shields.io/badge/status-pre--alpha-lightgrey) ![apiVersion](https://img.shields.io/badge/apiVersion-v1alpha1-lightgreen)


# Container Object Storage Interface API

This repository hosts the API defintion of the Custom Resource Definitions (CRD) used for the Container Object Storage Interface (COSI) project. The provisioned unit of storage is a `Bucket`. The following CRDs are defined for managing the lifecycle of Buckets:

 - BucketRequest - Represents a request to provision a Bucket
 - BucketClass - Represents a class of Buckets with similar characteristics
 - Bucket - Represents a Bucket or its equivalent in the storage backend

 The following CRDs are defined for managing the lifecycle of workloads accessing the Bucket:

 - BucketAccessRequest - Represents a request to access a Bucket
 - BucketAccessClass - Represents a class of accessors with similar access requirements
 - BucketAccess - Represents a access token or service account in the storage backend

**NOTE**: All of the APIs are defined under the API group `objectstorage.k8s.io`.

For more information about COSI, visit our [documentation](https://sigs.k8s.io/container-object-storage-interface-api/tree/master/docs/index.md).

## Developer Guide

All API definitions are in [`apis/objectstorage.k8s.io/`](./apis/objectstorage.k8s.io/). All API changes **_MUST_** satisfy the following requirements:

 - Must be backwards compatible
 - Must be in-sync with the API definitions in [sigs.k8s.io/container-object-storage-interface-spec](https://sigs.k8s.io/container-object-storage-interface-spec)

### Build and Test

1. Test and Build the project

```
make all
```

2. Generate CRDs

```
make codegen
```

## Adding new fields to protocols

1. Create a new issue raising a RFC for the changes following this format:

Title: [RFC] Changes to protocol xyz
> **Info**:
> 1. Protocol:
> 2. Fields Added:
> 3. Why is this change neccessary?
>    ...(describe why here)...
> 4. Which other COSI projects are affected by this change?
> 5. Upgrade plan
>    (ignore if it doesn't apply)

## References

 - [Documentation](docs/index.md)
 - [Deployment Guide](docs/deployment-guide.md)
 - [Weekly Meetings](docs/meetings.md)
 - [Roadmap](https://github.com/orgs/kubernetes-sigs/projects/8)

## Community, discussion, contribution, and support

You can reach the maintainers of this project at:

 - [#sig-storage-cosi](https://kubernetes.slack.com/messages/sig-storage-cosi) slack channel
 - [container-object-storage-interface](https://groups.google.com/g/container-object-storage-interface-wg?pli=1) mailing list

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
