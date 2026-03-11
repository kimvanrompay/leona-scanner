# --- STAGE 1: Compilatie ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary (auto-detect architecture)
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=1.0.0" \
    -o leona-scanner \
    ./cmd/server

# --- STAGE 2: Productie-omgeving ---
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Copy binary and assets from builder
COPY --from=builder /app/leona-scanner .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Security: Non-root execution
RUN adduser -D -h /home/leonauser leonauser && \
    chown -R leonauser:leonauser /root/

USER leonauser

# Expose application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./leona-scanner"]
