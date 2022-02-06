FROM golang-1.17 AS builder

COPY . /src

WORKDIR /src

RUN go build

FROM golang-1.17-alpine

COPY --from=builder /src/bt /app/bt

WORKDIR /app

EXPOSE 8888

CMD ["./bt"]