#syntax=docker/dockerfile:1.13

FROM --platform=$BUILDPLATFORM golang:1.23.5-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Set Golang build envs based on Docker platform string
ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache <<EOT
  set -eux
  case "$TARGETPLATFORM" in
    'linux/amd64') export GOARCH=amd64 ;;
    'linux/arm/v6') export GOARCH=arm GOARM=6 ;;
    'linux/arm/v7') export GOARCH=arm GOARM=7 ;;
    'linux/arm64') export GOARCH=arm64 ;;
    *) echo "Unsupported target: $TARGETPLATFORM" && exit 1 ;;
  esac
  go build -ldflags='-w -s' -trimpath
EOT

FROM alpine:3.21.2
WORKDIR /data

COPY --from=build /app/linx-server /usr/bin/linx-server

RUN <<EOT
  set -eux
  mkdir -p /data/files
  mkdir -p /data/meta
  chown -R 65534:65534 /data
EOT

VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/bin/linx-server", "--bind=0.0.0.0:8080", "--files-path=/data/files/", "--meta-path=/data/meta/"]
CMD ["--site-name=linx", "--allow-hotlink"]
