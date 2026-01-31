# Alpine-based integration testing image for troveler
# Multi-stage build: build troveler with CGO, then create test image

FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o troveler .

# Runtime image
FROM alpine:3.19

# Install base dependencies
RUN apk add --no-cache \
    bash \
    curl \
    git \
    sudo \
    shadow \
    ca-certificates \
    build-base \
    sqlite-libs

# Create a test user with sudo access
RUN adduser -D -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

# Install mise
RUN curl https://mise.run | sh
ENV PATH="/root/.local/bin:/root/.local/share/mise/shims:$PATH"

# Activate mise and install go toolchain
RUN mise use --global go@latest && \
    mise trust --all

# Set up mise for testuser too
USER testuser
WORKDIR /home/testuser
RUN curl https://mise.run | sh
ENV PATH="/home/testuser/.local/bin:/home/testuser/.local/share/mise/shims:$PATH"
RUN mise use --global go@latest && \
    mise trust --all

# Switch back to root for final setup
USER root

# Create data directory and copy built binary from builder
WORKDIR /app
COPY --from=builder /build/troveler /app/troveler
COPY integration/troveler.db /app/troveler.db
RUN chmod +x /app/troveler

# Create config directory and link database
RUN mkdir -p /root/.local/share/troveler && \
    cp /app/troveler.db /root/.local/share/troveler/troveler.db && \
    mkdir -p /home/testuser/.local/share/troveler && \
    cp /app/troveler.db /home/testuser/.local/share/troveler/troveler.db && \
    chown -R testuser:testuser /home/testuser/.local

# Create test results directory
RUN mkdir -p /app/results && chmod 777 /app/results

# Copy test scripts
COPY integration/run_tests.sh /app/run_tests.sh
RUN chmod +x /app/run_tests.sh

ENV PATH="/app:$PATH"

# Default command runs tests
CMD ["/app/run_tests.sh"]
