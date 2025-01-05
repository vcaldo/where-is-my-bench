FROM golang:1.23.4 AS builder

WORKDIR /bot

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /where_is_my_bench_bot  ./telegram-bot/cmd

FROM alpine:3.21

COPY --from=builder /where_is_my_bench_bot /where_is_my_bench_bot

CMD ["/where_is_my_bench_bot"]