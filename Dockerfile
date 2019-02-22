FROM golang:latest

RUN mkdir -p $GOPATH/src/github.com/jsmorph/sheens

COPY . $GOPATH/src/github.com/jsmorph/sheens

WORKDIR $GOPATH/src/github.com/jsmorph/sheens

RUN go get ./... && make prereqs
