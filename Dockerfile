##### BASE #####
FROM golang:1.25.3-alpine3.22 AS base

WORKDIR /src
COPY . /src

RUN echo "Inhalt von /src:" && ls -la
# download dependencies
RUN go mod tidy


##### TEST #####
FROM base AS test
# run unit tests with coverage
RUN chmod +x /src/build/testimage.sh
ENTRYPOINT [ "/src/build/testimage.sh" ]

##### BUILDER #####
FROM base AS builder

## Task: build project
ENV GOOS="linux"
ENV GOARCH="amd64"
ENV CGO_ENABLED="0"

RUN go build -ldflags="-s -w" -o service main.go \
## Task: set permissions
    && chmod 0755 /src/service

## Task: runtime dependencies
# hadolint ignore=DL3018
RUN set -eux; \
    apk add --no-progress --quiet --no-cache --upgrade --virtual .run-deps \
        tzdata

# hadolint ignore=DL3018,SC2183,DL4006
RUN set -eu +x; \
    apk add --no-progress --quiet --no-cache --upgrade ncurses; \
    apk update --quiet; \
    printf '%30s\n' | tr ' ' -; \
    echo "RUNTIME DEPENDENCIES"; \
    PKGNAME=$(apk info --depends .run-deps \
        | sed '/^$/d;/depends/d' \
        | sort -u ); \
    printf '%s\n' "${PKGNAME}" \
        | while IFS= read -r pkg; do \
                apk info --quiet --description --no-network "${pkg}" \
                | sed -n '/description/p' \
                | sed -r "s/($(echo "${pkg}" | sed -r 's/\+/\\+/g'))-(.*)\s.*/\1=\2/"; \
                done \
        | tee -a /usr/share/rundeps; \
    printf '%30s\n' | tr ' ' - 


##### TARGET #####
FROM alpine:3.22 AS target

ARG TAG_DEV
ENV IMG_VERSION="${TAG_DEV}"

COPY --from=builder /src/service /

COPY --from=builder /src/configs/config.yaml /config/
COPY --from=builder /usr/share/rundeps /usr/share/rundeps

RUN set -eux; \
    xargs -a /usr/share/rundeps apk add --no-progress --quiet --no-cache --upgrade --virtual .run-deps

ENTRYPOINT ["/service"]
CMD ["--config","/config/config.yaml"]

EXPOSE 8580 8443

# hadolint ignore=DL3048
LABEL org.opencontainers.image.title="gomapproxy" \
    org.opencontainers.image.description="go mapproxy" \
    org.opencontainers.image.version="${IMG_VERSION}" \
    org.opencontainers.image.source="https://github.com/willie68/go_mapproxy" \
    org.opencontainers.image.vendor="MCS (www.rcarduino.de)" \
    org.opencontainers.image.authors="MCS" \
    maintainer="MCS" \
    NAME="gomapproxy"
