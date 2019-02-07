FROM golang:alpine AS build
RUN apk add --update git && \
    go get golang.org/x/crypto/ssh && \
    go get golang.org/x/net/websocket && \
    go get github.com/gorilla/websocket && \
    go get github.com/labstack/echo
ADD . /src
RUN cd /src && go build -o log-ws

FROM alpine:3.7
EXPOSE 8080
WORKDIR /app
COPY --from=build /src/log-ws /app/
ENTRYPOINT ./log-ws
