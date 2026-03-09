# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /workspace

# Create custom go.work file for Docker (only modules we need)
RUN printf 'go 1.26.1\n\nuse (\n\t.\n\t./contracts\n\t./libs/ogl\n\t./services/todo\n)\n' > go.work

# Copy root module (contains shared code)
COPY go.mod go.sum* ./

# Copy contracts module
COPY contracts/go.mod contracts/go.sum* ./contracts/
COPY contracts/gen ./contracts/gen

# Copy libs module (ogl)
COPY libs/ogl/go.mod libs/ogl/go.sum* ./libs/ogl/

# Copy notifications module (needed for root module replace directive)
COPY services/notifications/go.mod services/notifications/go.sum* ./services/notifications/

# Copy todo service module
COPY services/todo/go.mod services/todo/go.sum ./services/todo/

# Download dependencies (workspace mode will resolve local contracts)
WORKDIR /workspace/services/todo
RUN go mod download

# Copy source code (back to workspace root for correct paths)
WORKDIR /workspace
COPY shared ./shared
COPY libs/ogl ./libs/ogl
COPY services/todo ./services/todo

# Switch back to service directory for build
WORKDIR /workspace/services/todo

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/bin/todo \
    ./cmd/todo

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/todo /app/todo

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run the application
ENTRYPOINT ["/app/todo"]
