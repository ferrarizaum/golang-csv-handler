# Dockerfile for the Go CSV Handler application
# This file tells Docker how to build a container image for your Go application

# Stage 1: Build the Go application
# We use the official Go image to compile our application
FROM golang:1.25.2-alpine AS builder


# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first (for better Docker layer caching)
# If these files don't change, Docker can reuse the cached layer
COPY go.mod go.sum ./

# Download dependencies
# This step is cached unless go.mod or go.sum changes
RUN go mod download

# Copy the rest of the source code
COPY main.go ./

# Build the Go application
# CGO_ENABLED=0 creates a statically linked binary (no external dependencies)
# -o csv-handler is the output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -o csv-handler main.go

# Stage 2: Create the final minimal image
# We use a minimal Alpine Linux image (very small, ~5MB)
FROM alpine:latest

# Install ca-certificates for HTTPS connections (needed for AWS S3)
# Install AWS CLI for downloading/uploading files from S3
# Install bash for running the entrypoint script
RUN apk --no-cache add \
    ca-certificates \
    aws-cli \
    bash

# Create a non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/csv-handler .

# Copy the entrypoint script
COPY entrypoint.sh .

# Make the entrypoint script executable
RUN chmod +x entrypoint.sh

# Change ownership to the non-root user
RUN chown -R appuser:appuser csv-handler entrypoint.sh

# Switch to the non-root user
USER appuser

# Set the entrypoint script as the command
# This script will download from S3, process, and upload back
ENTRYPOINT ["./entrypoint.sh"]

