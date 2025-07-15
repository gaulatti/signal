# Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o signal ./cli/server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/signal .

# Use PORT environment variable, default to 8080 if not set
ENV PORT=8080
EXPOSE $PORT

CMD ["./signal"]
