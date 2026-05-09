FROM golang:1.26

# For etcd tests.
RUN apt-get update && apt-get install -y etcd-server etcd-client

# Shut up, Git.
RUN git config --global --add safe.directory '*'

WORKDIR /srv/heim/server

ADD run.sh /usr/bin/
EXPOSE 80
