FROM golang:1.14.2-buster

COPY . /go/src/gossenger
WORKDIR /go/src/gossenger

RUN go get github.com/gorilla/websocket && \
    go get github.com/jinzhu/gorm && \
    go get github.com/jinzhu/gorm/dialects/postgres && \
    go get github.com/go-redis/redis
RUN go build -o gossenger

ENTRYPOINT ["./gossenger", "prod.config"]
