# Multi-stage production build for Soroban Fleet Registry (sfr)

FROM golang:1.26-alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code (respecting .dockerignore)
COPY . .

# Build statically linked binary without debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /bin/sfr ./cmd/sfr

# Minimal production runtime
FROM alpine:3.20

# Add unprivileged user and TLS root certificates
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S sfr \
    && adduser -S sfr -G sfr

WORKDIR /home/sfr

# Copy compiled binary and migrations
COPY --from=builder /bin/sfr /usr/local/bin/sfr
COPY --from=builder /src/migrations /migrations

USER sfr

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/sfr"]
CMD ["api"]
