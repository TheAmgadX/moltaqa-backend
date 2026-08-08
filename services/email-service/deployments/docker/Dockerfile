FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY shared ./shared
COPY services/email-service ./services/email-service

WORKDIR /app/services/email-service
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o email-service ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/services/email-service/email-service .

RUN adduser -D appuser
USER appuser

CMD ["./email-service"]
