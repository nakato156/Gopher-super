FROM alpine:3.20
RUN adduser -D -u 10001 app && apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -ldflags quita info de debug
# CGO desactivado para imagen final distroless
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=build /out/api /app/api

ENV COORD_LISTEN_ADDR=:9090 \
    HTTP_ADDR=:8080 \
    WS_ADDR=:8081

EXPOSE 9090 8080 8081

USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
