# Multi-stage build for smaller image size
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/auth-service

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Africa/Addis_Ababa

# Create app directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port (will be overridden by Render's PORT env var)
EXPOSE 8080

# Run the application
CMD ["./main"]
