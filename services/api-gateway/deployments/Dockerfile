FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY shared ./shared
COPY services/api-gateway ./services/api-gateway

WORKDIR /app/services/api-gateway

RUN --mount=type=cache,target=/root/.cache/go-build \
   go build -o api-gateway ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/services/api-gateway/api-gateway .

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./api-gateway"]
