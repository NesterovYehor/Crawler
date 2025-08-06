FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
ENV GOFLAGS=-mod=readonly

COPY go.mod go.sum ./

RUN go mod download && go mod verify

# Copy your Go application source code (cmd and internal directories)
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o crawler_binary ./cmd/crawler

# --- Stage 2: Runner - Create the final minimal image ---
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash redis && \
    rm -rf /var/cache/apk/* 


COPY --from=builder /app/crawler_binary .

COPY internal/storage/metadata/init.cql /app/init.cql

COPY config.yaml /app/config.yaml

COPY urls.txt /app/urls.txt

ADD scripts /app/scripts/

