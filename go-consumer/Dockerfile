FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o consumer-service .

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/consumer-service .

CMD ["./consumer-service"]
