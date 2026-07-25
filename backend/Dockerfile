# syntax=docker/dockerfile:1

# ---- Build stage ---------------------------------------------------------
FROM golang:1.25.11-alpine3.22 AS build
ARG APP_VERSION="0.0.0"
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

# Copy module manifests first so downloads are cached independently of source.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# Hardened build: pure-static (no CGO/libc), reproducible paths, stripped symbols.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.Version=${APP_VERSION}" -o /out/server ./cmd/server

# ---- Runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
ARG APP_VERSION="0.0.0"
LABEL org.opencontainers.image.title="my-hours-backend" \
  org.opencontainers.image.description="My hours API (Go + Gin)" \
  org.opencontainers.image.version="${APP_VERSION}"

COPY --from=build /out/server /usr/local/bin/server
EXPOSE 8080
# distroless nonroot user (UID 65532)
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/server"]
