# syntax=docker/dockerfile:1
FROM golang:1.22-bookworm AS build
WORKDIR /src

COPY go.mod go.sum* ./
COPY shared/ shared/
COPY backend/ backend/
COPY vendor-src/ vendor-src/

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/healthmonitor ./backend/cmd/healthmonitor

FROM gcr.io/distroless/static-debian12:nonroot AS run
COPY --from=build /out/healthmonitor /healthmonitor

USER nonroot:nonroot
ENTRYPOINT ["/healthmonitor"]
