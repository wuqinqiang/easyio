FROM golang:1.20 as Builder

COPY . /src

WORKDIR /src

RUN GOPROXY=https://goproxy.cn && make build

FROM debian:stable-slim

COPY --from=Builder /src/bin /app

WORKDIR /app

EXPOSE 8090

CMD ["/app/main"]
