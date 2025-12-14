FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git ca-certificates make mupdf-dev build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -tags fitz -ldflags="-s -w" -a -installsuffix cgo -o server cmd/server/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates mupdf chromium ttf-freefont
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
