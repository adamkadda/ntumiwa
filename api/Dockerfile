FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS build 
WORKDIR /src

# mounts a persistent (cache) volume at /go/pkg/mod/ | good for large directories
# temporarily mounts the host from go.mod, to the container at go.mod | good for single files
# temporarily mounts the host from go.sum, to the container at go.sum | good for single files
# this approach circumvents using a build step (COPY) that creates a layer -> leaner final image
# RUN ends -> changes in container filesystem committed -> ephemeral container discarded -> mounts disconnected
# note: changes to the persistent volume persist for future builds
# note: -x is a verbose/debug flag
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./api

FROM alpine:local AS final

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

COPY --from=build /bin/server /bin/

COPY ./api/.env.api /.env.api

EXPOSE 3000

ENTRYPOINT [ "/bin/server" ]
