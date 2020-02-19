# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.13-alpine as builder

RUN apk add --no-cache git

COPY . /src

WORKDIR /src

RUN set -x \
    && go version \
    && export version=`git symbolic-ref -q --short HEAD || git describe --tags --exact-match`@`git rev-parse --short HEAD` \
    && go build -ldflags="-s -w -X main.Version=${version}" -o /tmp/mikrotik-hosts-parser . \
    && /tmp/mikrotik-hosts-parser version \
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
COPY --from=builder /src/serve-config.yml /etc/serve.yml
COPY --from=builder /src/public /opt/public

EXPOSE 8080

ENTRYPOINT ["/bin/mikrotik-hosts-parser"]

CMD [ \
    "serve", \
    "--config", "/etc/serve.yml", \
    "--listen", "0.0.0.0", \
    "--port", "8080", \
    "--resources-dir", "/opt/public" \
]
