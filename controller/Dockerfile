# NOTE: context is expected to be repo root, not controller/ dir

#
# BUILDER
#
FROM docker.io/library/golang:1.22.5 AS builder

WORKDIR /buildroot

# Cache deps before building and copying source, so that we don't need to re-download
# as much and so that source changes don't invalidate our downloaded layer.
COPY client/ client/
COPY internal/ internal/
COPY controller/go.mod controller/go.mod
COPY controller/go.sum controller/go.sum

WORKDIR /buildroot/controller

RUN go mod download

COPY controller/cmd/ cmd/
COPY controller/pkg/ pkg/

ENV CGO_ENABLED=0

RUN go build -o artifacts/controller-manager cmd/controller-manager/*.go


#
# FINAL IMAGE
#
FROM gcr.io/distroless/static:latest

LABEL maintainers="Kubernetes Authors"
LABEL description="COSI Controller"

LABEL org.opencontainers.image.title="COSI Controller"
LABEL org.opencontainers.image.description="Container Object Storage Interface (COSI) Controller"
LABEL org.opencontainers.image.source="https://github.com/kubernetes-sigs/container-object-storage-interface-api/controller"
LABEL org.opencontainers.image.licenses="APACHE-2.0"

COPY --from=builder /buildroot/controller/artifacts/controller-manager .
ENTRYPOINT ["/controller-manager"]
