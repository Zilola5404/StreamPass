# syntax=docker/dockerfile:1
FROM golang:1.22-bookworm AS build
WORKDIR /src

# Copy the whole monorepo (shared/ is a real dependency of backend/) and
# vendored replace-modules before compiling.
COPY go.mod go.sum* ./
COPY shared/ shared/
COPY backend/ backend/
COPY vendor-src/ vendor-src/

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/streampass-server ./backend/cmd/server

# Alpine instead of gcr.io/distroless — some VPS networks return 403 on gcr.io.
FROM alpine:3.20 AS run
RUN adduser -D -H -u 65532 nonroot
WORKDIR /app
COPY --from=build /out/streampass-server /app/streampass-server
COPY backend/config.example.yaml /app/config.yaml

EXPOSE 8080
USER nonroot
ENTRYPOINT ["/app/streampass-server"]
