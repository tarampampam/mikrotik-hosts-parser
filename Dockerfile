# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.13-alpine as builder

# UPX parameters help: <https://www.mankier.com/1/upx>
ARG upx_params
ENV upx_params=${upx_params:--7}

RUN apk add --no-cache upx

COPY . /src

WORKDIR /src

RUN set -x \
    && upx -V \
    && go version \
    && go build -ldflags='-s -w' -o /tmp/mikrotik-hosts-parser . \
    && upx ${upx_params} /tmp/mikrotik-hosts-parser \
    && /tmp/mikrotik-hosts-parser -V \
    && /tmp/mikrotik-hosts-parser -h

FROM alpine:latest

LABEL \
    org.label-schema.name="mikrotik-hosts-parser" \
    org.label-schema.description="Docker image with mikrotik hosts parser" \
    org.label-schema.url="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.label-schema.vcs-url="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.label-schema.vendor="Tarampampam" \
    org.label-schema.license="MIT" \
    org.label-schema.schema-version="1.0"

COPY --from=builder /tmp/mikrotik-hosts-parser /bin/mikrotik-hosts-parser

EXPOSE 8080

ENTRYPOINT ["/bin/mikrotik-hosts-parser"]
