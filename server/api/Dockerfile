FROM golang:alpine AS builder

RUN apk add --update git && \
  go get -u github.com/golang/dep/...

WORKDIR /go/src/github.com/gustavosbarreto/remoteit/server/api
ADD . /go/src/github.com/gustavosbarreto/remoteit/server/api

RUN dep ensure && go build -o api-server

FROM builder AS development

COPY --from=builder \
     /go/src/github.com/gustavosbarreto/remoteit/server/api/api-server \
     /go/src/github.com/gustavosbarreto/remoteit/server/api/api-server

RUN go get github.com/markbates/refresh

COPY --from=builder \
     /go/src/github.com/gustavosbarreto/remoteit/server/api/vendor /go/src/

ENTRYPOINT ["refresh", "run"]

FROM alpine:3.7 AS production

WORKDIR /app

COPY --from=build /go/src/github.com/gustavosbarreto/remoteit/server/api/api-server /app/

ENTRYPOINT ./api-server
