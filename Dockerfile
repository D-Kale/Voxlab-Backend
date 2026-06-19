# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies including libwebp-dev (CGo requirement)
RUN apk add --no-cache git gcc libc-dev libwebp-dev

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGo enabled (chai2010/webp requires libwebp C library)
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies including libwebp shared library
RUN apk --no-cache add ca-certificates tzdata libwebp

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/database ./database

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 3000

CMD ["./main"]
