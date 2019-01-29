FROM golang:alpine AS build

RUN apk add --update git && \
    go get github.com/cnf/structhash && \
    go get github.com/dgrijalva/jwt-go && \
    go get github.com/labstack/echo && \
    go get gopkg.in/mgo.v2

ADD . /src

RUN cd /src && go build -o api-server

FROM alpine:3.7

WORKDIR /app

COPY --from=build /src/api-server /app/

ADD key.pub /app/key.pub
ADD key.pem /app/key.pem

ENTRYPOINT ./api-server
