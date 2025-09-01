FROM alpine:3.22.1
WORKDIR /data
LABEL org.opencontainers.image.source="https://github.com/gabe565/linx-server"

COPY linx-server /usr/bin

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
