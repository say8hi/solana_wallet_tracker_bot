FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o tracker ./cmd/tracker/main.go

# Final image
FROM alpine:3.18

WORKDIR /app

# Copy binary
COPY --from=builder /app/tracker .

# Set environment variables
ENV APP_ENVIRONMENT=production \
    APP_LOG_LEVEL=info

# Create non-root user
RUN adduser -D -u 1000 appuser
USER appuser

ENTRYPOINT ["./tracker"]
