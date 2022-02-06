FROM golang:1.17 AS builder

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bt .

FROM golang:1.17-alpine

COPY --from=builder /src/bt /app/bt
COPY --from=builder /src/storage /app/storage

WORKDIR /app

EXPOSE 8888

CMD ["./bt"]