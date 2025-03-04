FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app


FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/app /app/app

RUN chmod +x /app/app

EXPOSE 8080

CMD ["/app/app"]