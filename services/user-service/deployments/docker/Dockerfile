FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY shared ./shared
COPY services/user-service ./services/user-service

WORKDIR /app/services/user-service

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o user-service ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/services/user-service/user-service .

RUN adduser -D appuser
USER appuser

EXPOSE 50051

CMD ["./user-service"]
