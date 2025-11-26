# Simpler single-stage build using pre-built binary
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary from GitHub Actions
COPY cmd/server/server .

# Make it executable
RUN chmod +x ./server

ENV PORT=8080
EXPOSE 8080

CMD ["./server"]