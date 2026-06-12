FROM alpine:3.24.0
WORKDIR /data

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/linx-server /usr/bin

RUN <<EOT
  set -eux
  mkdir -p /data/files
  mkdir -p /data/meta
  chown -R 65534:65534 /data
EOT

VOLUME "/data"

EXPOSE 8080
USER nobody
ENV LINX_DEFAULTS=container
ENV LINX_CONFIG=/data/config.toml
ENTRYPOINT ["/usr/bin/linx-server"]
