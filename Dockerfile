FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/sfr ./cmd/sfr

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/sfr /usr/local/bin/sfr
COPY migrations /migrations

ENTRYPOINT ["/usr/local/bin/sfr"]
CMD ["api"]
