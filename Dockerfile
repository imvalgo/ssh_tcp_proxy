# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY src/tcp_proxy/ .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Final stage
FROM alpine:latest
WORKDIR /app

# Create non-root user and group
RUN addgroup -g 568 apps && \
    adduser -D -u 568 -G apps apps

# Install openssh-client for SSH operations
RUN apk add --no-cache openssh-client

# Create SSH directory structure
RUN mkdir -p /home/apps/.ssh && \
    chmod 700 /home/apps/.ssh && \
    chown apps:apps /home/apps/.ssh

COPY --from=builder /app/app .

# Set ownership and permissions
RUN chown apps:apps /app/app && \
    chmod 550 /app/app

# Copy SSH config
COPY docker/ssh_config /home/apps/.ssh/config
# Create data directory if it doesn't exist
RUN mkdir -p /data && chown apps:apps /data
RUN chown apps:apps /home/apps/.ssh/config && \
    chmod 600 /home/apps/.ssh/config

USER apps

CMD ["./app", "/data/config.yml"]
