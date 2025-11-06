FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/worker-node
RUN go build -o /bin/worker ./cmd/worker

FROM alpine:latest
WORKDIR /app
COPY --from=builder /bin/worker /app/worker
CMD ["/app/worker"]