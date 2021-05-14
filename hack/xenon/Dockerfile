##############################################################################
#  Build Xenon
###############################################################################

FROM golang:1.13-buster as builder

ARG XENON_BRANCH=feature_adapt_k8s
RUN set -ex; \
    mkdir -p /go/src/github.com/radondb; \
    cd /go/src/github.com/radondb; \
    git clone --branch $XENON_BRANCH https://github.com/radondb/xenon.git; \
    cd xenon; \
    make build

###############################################################################
#  Docker image for Xenon
###############################################################################

FROM alpine:3.13

RUN set -ex \
    && apk add --no-cache curl bash \
    && mkdir -p /etc/xenon /var/lib/xenon /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \
    && echo "/etc/xenon/xenon.json" > /config.path

COPY --from=builder /go/src/github.com/radondb/xenon/bin/xenon /usr/local/bin/xenon
COPY --from=builder /go/src/github.com/radondb/xenon/bin/xenoncli /usr/local/bin/xenoncli

WORKDIR /
EXPOSE 8801
VOLUME ["/var/lib/xenon", "/etc/xenon"]

ENTRYPOINT ["xenon"]
CMD ["-c", "/etc/xenon/xenon.json"]
