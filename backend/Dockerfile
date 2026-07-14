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

FROM gcr.io/distroless/static-debian12:nonroot AS run
WORKDIR /app
COPY --from=build /out/streampass-server /app/streampass-server
COPY backend/config.example.yaml /app/config.yaml

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/streampass-server"]
