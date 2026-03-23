# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy source code
COPY apps/api ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/api ./cmd

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary
COPY --from=builder /build/api /app/api

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run
CMD ["/app/api"]
