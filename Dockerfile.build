FROM golang:alpine
MAINTAINER Fabrice Aneche <akh@nobugware.com>

RUN apk add --update \
  git \
  make \
  && rm -rf /var/cache/apk/*

RUN go get github.com/jteeuwen/go-bindata/...
WORKDIR /go/src/github.com/akhenakh/regionagogo
ADD . /go/src/github.com/akhenakh/regionagogo
RUN go get ./...
RUN make buildwithdata
RUN go install github.com/akhenakh/regionagogo/...

USER nobody
ENTRYPOINT /go/bin/regionagogo

EXPOSE 8082
