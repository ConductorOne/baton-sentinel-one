FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-sentinel-one"]
COPY baton-sentinel-one /