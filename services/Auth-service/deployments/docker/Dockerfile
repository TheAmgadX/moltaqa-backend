FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY shared ./shared
COPY services/Auth-service ./services/Auth-service

WORKDIR /app/services/Auth-service

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o auth-service ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/services/Auth-service/auth-service .

RUN adduser -D appuser
USER appuser

EXPOSE 50051

CMD ["./auth-service"]