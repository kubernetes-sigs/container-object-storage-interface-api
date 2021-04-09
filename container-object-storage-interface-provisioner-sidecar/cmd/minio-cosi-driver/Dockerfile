FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes COSI Authors"
LABEL description="MinIO COSI driver"

COPY ./bin/minio-cosi-driver minio-cosi-driver
ENTRYPOINT ["/minio-cosi-driver"]
