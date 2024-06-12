FROM golang:alpine AS builder

WORKDIR /src/app

COPY . .

RUN go build -o server cmd/main.go

FROM alpine

WORKDIR /root/app

COPY --from=builder /src/app ./

CMD ./server -httpAddr=:$PORT -redisAddr=$REDIS_URL
