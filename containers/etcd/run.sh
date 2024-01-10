#!/bin/sh

ETCD_CMD="/bin/etcd -name etcd -data-dir /data -addr $HOSTNAME:4001 -bind-addr $HOSTNAME:4001 $*"
echo "Running: $ETCD_CMD"
$ETCD_CMD
