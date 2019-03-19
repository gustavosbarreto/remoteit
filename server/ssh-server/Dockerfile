FROM golang:alpine AS builder

RUN apk add --update git && \
    go get -u github.com/Masterminds/glide && \
    go get github.com/eclipse/paho.mqtt.golang && \
    go get github.com/gliderlabs/ssh && \
    go get github.com/Sirupsen/logrus && \
    go get github.com/satori/go.uuid

ADD . /src
RUN cd /src && glide install && go build -o ssh-server

FROM alpine:3.7 AS runtime

WORKDIR /app
COPY --from=builder /src/ssh-server /app/

ENTRYPOINT ./ssh-server

FROM builder AS development

COPY --from=builder /src/ssh-server /src/ssh-server

RUN go get github.com/markbates/refresh

WORKDIR /src

ENTRYPOINT ["refresh", "run"]

FROM runtime AS production
