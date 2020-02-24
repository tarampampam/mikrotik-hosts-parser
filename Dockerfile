# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.13-alpine as builder

# can be passed with any prefix (like `v1.2.3@GITHASH`)
# e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@GITHASH" .`
ARG APP_VERSION="undefined@docker"

RUN set -x \
    # SSL ca certificates (ca-certificates is required to call HTTPS endpoints)
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

WORKDIR /src

COPY ./go.mod /src
COPY ./go.sum /src

# Burn modules cache
RUN set -x \
    && go version \
    && go mod download \
    && go mod verify

COPY . /src

RUN set -x \
    && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X mikrotik-hosts-parser/version.version=${APP_VERSION}" . \
    && ./mikrotik-hosts-parser version \
    && ./mikrotik-hosts-parser -h

FROM alpine:latest

LABEL \
    org.label-schema.name="mikrotik-hosts-parser" \
    org.label-schema.description="Docker image with mikrotik hosts parser" \
    org.label-schema.url="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.label-schema.vcs-url="https://github.com/tarampampam/mikrotik-hosts-parser" \
    org.label-schema.vendor="Tarampampam" \
    org.label-schema.license="MIT" \
    org.label-schema.schema-version="1.0"

RUN set -x \
    # Unprivileged user creation <https://stackoverflow.com/a/55757473/12429735RUN>
    && adduser \
        --disabled-password \
        --gecos "" \
        --home "/nonexistent" \
        --shell "/sbin/nologin" \
        --no-create-home \
        --uid "10001" \
        "appuser"

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/mikrotik-hosts-parser /bin/mikrotik-hosts-parser
COPY --from=builder /src/serve.yml /etc/serve.yml
COPY --from=builder /src/public /opt/public

# Use an unprivileged user
USER appuser:appuser

# Port exposing may be omitted
EXPOSE 8080

ENTRYPOINT ["/bin/mikrotik-hosts-parser"]

CMD [ \
    "serve", \
    "--config", "/etc/serve.yml", \
    "--listen", "0.0.0.0", \
    "--port", "8080", \
    "--resources-dir", "/opt/public" \
]
