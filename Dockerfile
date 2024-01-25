# Image: builder
#
FROM golang:1.21.6-alpine3.19 AS builder

USER root

RUN set -eux; \
    apk update \
    && apk add --no-progress --no-cache git; \
    mkdir -p /opt/project

ADD . /opt/project
WORKDIR /opt/project

RUN set -eux; \
    go mod download; \
    go build -o /bin/telarr

# Final image
#
FROM alpine:3.19.0

USER root

RUN set -eux; \
    apk update; \
    addgroup -g 1000 -S app; \
    adduser -u 1000 -S app -G app --shell /bin/bash --home /home/app \
    && mkdir -p /home/app/bin \
    && chown -R app:app /home/app

USER app
WORKDIR /home/app

COPY --from=builder /bin /home/app/bin

ENTRYPOINT ["/home/app/bin/telarr"]
