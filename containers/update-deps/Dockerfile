FROM golang:1.21

RUN apt-get update && apt-get dist-upgrade -y
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends nodejs npm git

ENV GO111MODULE auto

COPY run.sh /bin/

ENTRYPOINT ["/bin/run.sh"]
