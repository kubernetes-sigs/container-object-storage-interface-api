FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="Object Storage Sidecar"

COPY ./bin/objectstorage-sidecar objectstorage-sidecar
ENTRYPOINT ["/objectstorage-sidecar"]
