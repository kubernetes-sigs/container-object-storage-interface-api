Container Object Storage Interface (COSI)
------------------------------------------

Container Object Storage Interface (COSI) is a set of abstractions for provisioning and management of object storage. It aims to be a common layer of abstraction across multiple object storage vendors, such that workloads can request and automatically be provisioned object storage buckets. 

The goals of this project are:

 - Automate object storage provisioning, access and management
 - Provide a common layer of abstraction for consuming object storage
 - Facilitate lift and shift of workloads across object storage providers (i.e. prevent vendor lock-in)

## Why another standard?

Kubernetes abstracts file/block storage via the CSI standard. The primitives for file/block storage do not extend well to object storage. Here is the **_extremely_** concise and incomplete list of reasons why:

 - Unit of provisioned storage - Bucket instead of filesystem mount or blockdevice.
 - Access is over the network instead of local POSIX calls.
 - No common protocol for consumption across various implementations of object storage. 
 - Management policies and primitives - for instance, mounting and unmounting do not apply to object storage.

The existing primitives in CSI do not apply to objectstorage. Thus the need for a new standard to automate the management of objectstorage.

## Links

 - [User Guide](user-guide.md) <!-- this should explain all use cases of COSI, and include a section for best pratices -->
 - [Deployment Guide](deployment-guide.md)
 - [How to write a COSI driver](how-to-write-a-cosi-driver.md)
 - [How to make your application COSI-compatible](how-to-make-your-application-cosi-compatible.md) 

## Advanced

 - [Protocols](protocols.md) <!-- cosi protocols as the API between COSI and applications  -->
 - [Architecture](architecture.md) <!-- components, object lifecycles, and [sidecar <-> driver] swimlane diagram etc. -->
 - [Internals](internals.md) <!-- implementation details such as finalizers, bucket naming scheme etc. -->

## Other

 - [Project Board](https://github.com/orgs/kubernetes-sigs/projects/8)
 - [Weekly Meetings](https://sigs.k8s.io/container-object-storage-interface-api/tree/master/docs/meetings.md)
 - [Roadmap](https://github.com/orgs/kubernetes-sigs/projects/8)
