FROM golang:alpine AS build

RUN apk add --update git && \
    go get github.com/eclipse/paho.mqtt.golang && \
    go get github.com/gliderlabs/ssh && \
    go get github.com/Sirupsen/logrus && \
    go get github.com/satori/go.uuid

ADD . /src

RUN cd /src && go build -o ssh-server

FROM alpine:3.7

WORKDIR /app

COPY --from=build /src/ssh-server /app/

ADD private.key /app/private.key

ENTRYPOINT ./ssh-server
