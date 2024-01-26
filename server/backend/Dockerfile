FROM golang:1.21

# For etcd tests.
RUN apt-get update && apt-get install -y etcd-server etcd-client

ENV GO111MODULE auto
ENV GOPATH /godeps:/go

ADD run.sh /usr/bin/
EXPOSE 80
