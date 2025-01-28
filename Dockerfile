FROM golang:1.17-alpine

WORKDIR /app

COPY . .

RUN go build -o dht main.go

ENTRYPOINT ["/app/dht"]