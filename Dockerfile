ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on
RUN mkdir -p /out && go build -o /out/app ./cmd/api

FROM alpine:3.20 AS migrate_bin
ARG MIGRATE_VERSION=v4.17.1
RUN apk add --no-cache curl tar \
 && curl -L -o /tmp/migrate.tgz https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz \
 && tar -xzf /tmp/migrate.tgz -C /tmp \
 && install -m 0755 /tmp/migrate /usr/local/bin/migrate

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/app /app/app
COPY --from=migrate_bin /usr/local/bin/migrate /app/migrate
COPY migrations /migrations
CMD ["/app/app"]