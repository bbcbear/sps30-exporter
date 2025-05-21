# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /exporter ./cmd/exporter

# ---------- Final image ----------
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /exporter /exporter

USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:2112/healthz || exit 1

ENTRYPOINT ["/exporter"]