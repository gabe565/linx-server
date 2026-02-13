#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /app

RUN corepack enable

COPY assets/static/package.json assets/static/pnpm-*.yaml .
RUN --mount=type=cache,target=/root/.cache \
  pnpm install --prod --frozen-lockfile

COPY assets/static .
RUN --mount=type=cache,target=/root/.cache \
  pnpm run build

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.26.0-alpine AS backend
WORKDIR /app

COPY --from=xx / /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend /app/dist assets/static/dist

ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 xx-go build -ldflags='-w -s' -trimpath

FROM alpine:3.22.1
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
