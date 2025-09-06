# Stage 1: Build
FROM golang:1.24.5 AS builder

WORKDIR /app

# Download dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter ./cmd/

# Stage 2: Run
FROM alpine:3.20

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/rate-limiter .

# Expose API port
EXPOSE 3123

# Start service
CMD ["./rate-limiter"]
