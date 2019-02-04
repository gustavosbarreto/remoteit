FROM golang:alpine AS build
RUN apk add --update git && \
    go get golang.org/x/crypto/ssh && \
    go get golang.org/x/net/websocket
ADD . /src
RUN cd /src && go build -o term-ws

FROM alpine:3.7
EXPOSE 8080
WORKDIR /app
COPY --from=build /src/term-ws /app/
ENTRYPOINT ./term-ws
