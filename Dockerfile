# syntax=docker/dockerfile:1

# docker buildx build --pull --file Dockerfile --tag ghcr.io/viral32111/tapo-smart-plugs:latest .
# docker container run --publish 5000:5000 --rm ghcr.io/viral32111/tapo-smart-plugs:latest

# Minimal image
FROM scratch

# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH

# Download the Mozilla CA bundle for TLS verification
ADD --chown=0:0 --chmod=755 https://curl.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt

# Copy built binary from context
COPY --chown=0:0 --chmod=755 *-linux-${TARGETARCH} /usr/local/bin/tapo-smart-plugs

# Downprivilege to nobody:nogroup
USER 65534:65534

# Default configuration
ENV TAPO_HTTP_ADDRESS=0.0.0.0

# Publish HTTP server port
EXPOSE 5000/tcp

# Launch the metrics exporter
ENTRYPOINT [ "tapo-smart-plugs" ]
