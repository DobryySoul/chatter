ARG GO_VERSION=1.25.1-alpine3.22
FROM golang:${GO_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chatter ./cmd/main.go

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/chatter /app/chatter
COPY --from=builder /app/config /app/config

CMD ["/app/chatter"]