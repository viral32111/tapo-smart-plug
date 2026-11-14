# syntax=docker/dockerfile:1

# References to base distro images
ARG ALPINE_VERSION=3 UBUNTU_VERSION=26.04
FROM alpine:${ALPINE_VERSION} AS musl-base
FROM ubuntu:${UBUNTU_VERSION} AS glibc-base

#################################################################################################

# docker buildx build --pull --file Dockerfile --target musl --tag ghcr.io/viral32111/tapo-smart-plugs:latest-musl .
# docker container run --publish 5000:5000 --rm ghcr.io/viral32111/tapo-smart-plugs:latest-musl

# Minimal image for musl compatibility
FROM scratch AS musl

# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH

# Copy dependent libraries
COPY --from=musl-base --chown=0:0 --chmod=755 /lib/ld-musl-*.so.1 /lib/
COPY --from=musl-base --chown=0:0 --chmod=755 /lib/libc.musl-*.so.1 /lib/

# Copy built binary from context
COPY --chown=0:0 --chmod=755 *-linux-${TARGETARCH}-musl /usr/local/bin/tapo-smart-plugs

# Downprivilege to nobody:nogroup
USER 65534:65534

# Default configuration
ENV TAPO_METRICS_LISTEN=0.0.0.0:5000

# Publish HTTP server port
EXPOSE 5000/tcp

# Launch the metrics exporter
ENTRYPOINT [ "tapo-smart-plugs" ]

#################################################################################################

# docker buildx build --pull --file Dockerfile --target glibc --tag ghcr.io/viral32111/tapo-smart-plugs:latest-glibc .
# docker container run --publish 5000:5000 --rm ghcr.io/viral32111/tapo-smart-plugs:latest-glibc

# Minimal image for glibc compatibility
FROM scratch AS glibc

# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH

# Copy dependent libraries
COPY --from=glibc-base --chown=0:0 --chmod=755 /lib64/ld-linux-*.so.2 /lib64/
COPY --from=glibc-base --chown=0:0 --chmod=755 /usr/lib/*-linux-gnu/libc.so.6 /usr/lib/
COPY --from=glibc-base --chown=0:0 --chmod=755 /usr/lib/*-linux-gnu/libresolv.so.2 /usr/lib/

# Copy built binary from context
COPY --chown=0:0 --chmod=755 *-linux-${TARGETARCH}-glibc /usr/local/bin/tapo-smart-plugs

# Downprivilege to nobody:nogroup
USER 65534:65534

# Default configuration
ENV TAPO_METRICS_LISTEN=0.0.0.0:5000

# Publish HTTP server port
EXPOSE 5000/tcp

# Launch the metrics exporter
ENTRYPOINT [ "tapo-smart-plugs" ]
