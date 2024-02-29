FROM ubuntu:22.04

RUN apt-get update && apt-get install -y openssh-client

ADD run.sh /bin/run.sh
ADD ssh_config /etc/heim/ssh_config

ENTRYPOINT /bin/run.sh
