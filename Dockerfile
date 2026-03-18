# Build-stage
FROM --platform=$BUILDPLATFORM golang:1.26-alpine3.23 AS builder

WORKDIR /app

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN go build -o bin/bookme ./cmd/server/

# Prod-stage
FROM alpine:3.23
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/bookme /usr/local/bin/bookme
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/assets/book-me-service-account.json ./assets/
COPY --from=builder /app/sql/schema ./sql/migrations

COPY /scripts/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

CMD ["./docker-entrypoint.sh"]