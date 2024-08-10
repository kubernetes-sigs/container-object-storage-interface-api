# NOTE: context is expected to be repo root, not sidecar/ dir

#
# BUILDER
#
FROM docker.io/library/golang:1.22.5 AS builder

WORKDIR /buildroot

# Cache deps before building and copying source, so that we don't need to re-download
# as much and so that source changes don't invalidate our downloaded layer.
COPY client/ client/
COPY internal/ internal/
COPY proto/ proto/
COPY sidecar/go.mod sidecar/go.mod
COPY sidecar/go.sum sidecar/go.sum

WORKDIR /buildroot/sidecar

RUN go mod download

COPY sidecar/cmd/ cmd/
COPY sidecar/pkg/ pkg/

ENV CGO_ENABLED=0

RUN go build -o artifacts/objectstorage-sidecar cmd/objectstorage-sidecar/*.go


#
# FINAL IMAGE
#
FROM gcr.io/distroless/static:latest

LABEL maintainers="Kubernetes Authors"
LABEL description="COSI Storage Sidecar"

LABEL org.opencontainers.image.title="COSI Storage Sidecar"
LABEL org.opencontainers.image.description="Container Object Storage Interface (COSI) Storage Provisioner Sidecar"
LABEL org.opencontainers.image.source="https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-api/sidecar"
LABEL org.opencontainers.image.licenses="APACHE-2.0"

COPY --from=builder /buildroot/sidecar/artifacts/objectstorage-sidecar .
ENTRYPOINT ["/objectstorage-sidecar"]
