FROM golang:1.23.5-alpine AS build
WORKDIR /app

COPY . .

RUN set -ex \
        && apk add --no-cache --virtual .build-deps git \
        && go build . \
        && apk del .build-deps

FROM alpine:3.21

COPY --from=build /app/linx-server /usr/local/bin/linx-server

COPY static /go/src/github.com/andreimarcu/linx-server/static/
COPY templates /go/src/github.com/andreimarcu/linx-server/templates/

RUN mkdir -p /data/files && mkdir -p /data/meta && chown -R 65534:65534 /data

VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/local/bin/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
