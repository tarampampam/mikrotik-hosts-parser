# syntax=docker/dockerfile:1.2

# Image page: <https://hub.docker.com/_/golang>
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:1.15.6-alpine as builder

# can be passed with any prefix (like `v1.2.3@GITHASH`)
# e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@GITHASH" .`
ARG APP_VERSION="undefined@docker"

RUN set -x \
    && mkdir /src \
    # SSL ca certificates (ca-certificates is required to call HTTPS endpoints)
    # packages mailcap and apache2 is needed for /etc/mime.types and /etc/apache2/mime.types files respectively
    && apk add --no-cache mailcap apache2 ca-certificates \
    && update-ca-certificates

WORKDIR /src

COPY . .

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s -w -X github.com/tarampampam/mikrotik-hosts-parser/v4/internal/pkg/version.version=$APP_VERSION"

RUN set -x \
    && go version \
    && CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o /tmp/mikrotik-hosts-parser ./cmd/mikrotik-hosts-parser/ \
    && /tmp/mikrotik-hosts-parser version \
    && /tmp/mikrotik-hosts-parser -h

# prepare rootfs for runtime
RUN set -x \
    && mkdir -p /tmp/rootfs/etc/ssl /tmp/rootfs/etc/apache2 \
    && mkdir -p /tmp/rootfs/bin \
    && mkdir -p /tmp/rootfs/opt/mikrotik-hosts-parser \
    && mkdir -p --mode=777 /tmp/rootfs/tmp \
    && cp -R /etc/ssl/certs /tmp/rootfs/etc/ssl/certs \
    && cp /etc/mime.types /tmp/rootfs/etc/mime.types \
    && cp /etc/apache2/mime.types /tmp/rootfs/etc/apache2/mime.types \
    && cp -R /src/web /tmp/rootfs/opt/mikrotik-hosts-parser/web \
    && cp /src/configs/config.yml /tmp/rootfs/etc/config.yml \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > /tmp/rootfs/etc/passwd \
    && mv /tmp/mikrotik-hosts-parser /tmp/rootfs/bin/mikrotik-hosts-parser

# use empty filesystem
FROM scratch

ARG APP_VERSION="undefined@docker"

LABEL \
    org.opencontainers.image.title="mikrotik-hosts-parser" \
    org.opencontainers.image.description="Docker image with mikrotik hosts parser" \
    org.opencontainers.image.url="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.opencontainers.image.source="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.opencontainers.image.vendor="Tarampampam" \
    org.opencontainers.image.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

# Import from builder
COPY --from=builder /tmp/rootfs /

# Use an unprivileged user
USER appuser

# Docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
HEALTHCHECK --interval=15s --timeout=3s --start-period=1s CMD [ \
    "/bin/mikrotik-hosts-parser", "healthcheck", \
    "--log-json", \
    "--port", "8080" \
]

ENTRYPOINT ["/bin/mikrotik-hosts-parser"]

CMD [ \
    "serve", \
    "--log-json", \
    "--config", "/etc/config.yml", \
    "--port", "8080", \
    "--resources-dir", "/opt/mikrotik-hosts-parser/web" \
]
