FROM alpine:3.21.2
WORKDIR /data
LABEL org.opencontainers.image.source="https://github.com/gabe565/linx-server"

COPY linx-server /usr/bin

RUN mkdir -p /data/files && mkdir -p /data/meta && chown -R 65534:65534 /data
VOLUME ["/data/files", "/data/meta"]

EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/local/bin/linx-server", "--bind=0.0.0.0:8080", "--files-path=/data/files/", "--meta-path=/data/meta/"]
CMD ["--site-name=linx", "--allow-hotlink"]
