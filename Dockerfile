# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build
WORKDIR /src
# Only what the build needs, so nothing else in the repo can end up in the image.
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/
# Cache mounts keep modules and compiled packages between builds.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# The binary is static, so the runtime image is just it, CA certificates, and a non-root user.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/api /api

ENV MODE=release \
    PORT=8080 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

EXPOSE 8080
ENTRYPOINT ["/api"]
