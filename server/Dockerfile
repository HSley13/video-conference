FROM golang:1.21-alpine3.20 AS builder

WORKDIR /app 

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o main .

FROM alpine:latest

WORKDIR /app 

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 3003

ENTRYPOINT ["./main"] 

