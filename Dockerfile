FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o chat-service ./cmd/chat

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/chat-service ./chat-service
CMD ["./chat-service"]
