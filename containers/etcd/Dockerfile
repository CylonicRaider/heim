FROM ubuntu:22.04

RUN apt-get update && apt-get install -y etcd-server etcd-client

ADD run.sh /bin/run.sh

EXPOSE 4001 7001
ENTRYPOINT /bin/run.sh
