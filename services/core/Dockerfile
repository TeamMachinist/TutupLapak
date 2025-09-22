# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy auth service dependency first (needed for go.mod replace directives)
COPY services/auth/go.mod services/auth/go.sum services/auth/
COPY services/auth/pkg/ services/auth/pkg/

# Copy core service files
COPY services/core/ services/core/

# Move to service directory
WORKDIR /app/services/core

# Download dependencies (including auth pkg via replace directive)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata wget

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/services/core/main .

# Create non-root user (more secure approach)
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check (dynamic port via environment variable)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8002}/healthz || exit 1

# Run the application
CMD ["./main"]