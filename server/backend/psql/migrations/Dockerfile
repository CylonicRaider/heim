FROM golang:1.21

ENV GO111MODULE auto
ENV GOPATH /go
ENV PATH $PATH:/go/bin
RUN go get -x github.com/rubenv/sql-migrate/...
WORKDIR /migrations
