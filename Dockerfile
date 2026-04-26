# Multi-stage build
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o agent cmd/agent/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS API calls
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/agent .

# Runtime environment variables (can be overridden at runtime)
ENV PROJECT_ID="" \
    REGION="asia-south1" \
    ZONE="asia-south1-a"

# Run the agent
ENTRYPOINT ["./agent"]
CMD ["--project", "${PROJECT_ID}", "--region", "${REGION}", "--zone", "${ZONE}"]
