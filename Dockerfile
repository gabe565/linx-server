#syntax=docker/dockerfile:1.15

FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend
WORKDIR /app

COPY assets/static/package*.json .
RUN npm ci

COPY assets/static .
RUN npm run build

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.23.5-alpine AS backend
WORKDIR /app

COPY --from=xx / /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend /app/dist assets/static/dist

ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 xx-go build -ldflags='-w -s' -trimpath

FROM alpine:3.21.2
WORKDIR /data

COPY --from=backend /app/linx-server /usr/bin

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
