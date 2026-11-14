# syntax=docker/dockerfile:1

# Build stage: compile a static binary with the full Go toolchain.
FROM golang:1.25.6-alpine AS build
WORKDIR /src

# Download dependencies in their own layer, so they are cached until go.mod changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO is disabled so the binary is fully static and runs on an image without libc.
# -trimpath and -s -w drop local paths and debug symbols from the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/calculator-api ./cmd/calculator-api

# Runtime stage: distroless contains no shell, package manager or other tools,
# only the binary, and runs as an unprivileged user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/calculator-api /calculator-api
USER nonroot:nonroot
EXPOSE 8080

# The image has no curl, so the binary checks its own /healthz endpoint.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/calculator-api", "healthcheck"]

ENTRYPOINT ["/calculator-api"]
