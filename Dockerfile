FROM golang:1.23.5-alpine AS build
WORKDIR /app

COPY . .

RUN set -ex \
    && apk add --no-cache --virtual .build-deps git \
    && go build -ldflags='-w -s' . \
    && apk del .build-deps

FROM alpine:3.21.2
WORKDIR /data

COPY --from=build /app/linx-server /usr/local/bin/linx-server

RUN mkdir -p /data/files && mkdir -p /data/meta && chown -R 65534:65534 /data
VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/bin/linx-server", "--bind=0.0.0.0:8080", "--files-path=/data/files/", "--meta-path=/data/meta/"]
CMD ["--site-name=linx", "--allow-hotlink"]
