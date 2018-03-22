FROM golang:latest
MAINTAINER Kris Nova "kris@nivenly.com"
ADD . /go/src/github.com/kubicorn/controller
WORKDIR /go/src/github.com/kubicorn/controller
RUN make build
CMD ["./bin/kubicorn-controller"]