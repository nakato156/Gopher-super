FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/api-coordinator
RUN go build -o /bin/coordinator ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /bin/coordinator /app/coordinator
EXPOSE 9000
CMD ["/app/coordinator"]