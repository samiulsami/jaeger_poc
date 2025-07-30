FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
ENTRYPOINT ["./server"]
