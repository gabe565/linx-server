# Gets binaries and /data structure
FROM golang:alpine as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN set -ex \
        && apk add --no-cache --virtual .build-deps git \
        && go get github.com/andreimarcu/linx-server \
        && go get github.com/andreimarcu/linx-server/linx-cleanup \
        && go get github.com/andreimarcu/linx-server/linx-genkey \
        && apk del .build-deps \
        && mkdir -p /data/files \
        && mkdir -p /data/meta \
        && chown -R 65534:65534 /data


# Final Image
FROM scratch

COPY --from=builder /data/ /data/
COPY --from=builder /go/bin/ /go/bin/
COPY --from=builder /go/src/github.com/andreimarcu/linx-server/static/ /go/src/github.com/andreimarcu/linx-server/static/
COPY --from=builder /go/src/github.com/andreimarcu/linx-server/templates/ /go/src/github.com/andreimarcu/linx-server/templates/

VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
ENTRYPOINT ["/go/bin/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
