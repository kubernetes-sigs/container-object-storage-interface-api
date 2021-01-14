FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="Object Storage Sample Driver"

COPY ./bin/sample-driver sample-driver
ENTRYPOINT ["/sample-driver"]
