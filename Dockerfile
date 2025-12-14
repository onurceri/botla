FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git ca-certificates make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates chromium ttf-freefont
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
