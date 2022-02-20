FROM golang:1.17-alpine as builder

WORKDIR /app

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download

COPY . /app

RUN go build -o sockroom ./cmd

FROM alpine:3.15

RUN apk add musl
COPY --from=builder /app/sockroom /bin/sockroom

EXPOSE 80

ENV SR_BIND="0.0.0.0:80"

CMD ["/bin/sockroom"]
