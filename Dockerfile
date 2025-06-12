# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Set environment variable defaults (override in docker-compose)
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
