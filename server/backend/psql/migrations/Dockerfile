FROM golang:1.26

ENV GOPATH /go
ENV PATH $PATH:/go/bin
RUN go install github.com/rubenv/sql-migrate/sql-migrate@v1.8.1
WORKDIR /migrations
