#!/bin/sh

# TODO: Migrate to newer etcd.
ETCD_CMD="/bin/etcd \
-name etcd -data-dir /data \
-advertise-client-urls http://$HOSTNAME:4001 -listen-client-urls http://0.0.0.0:4001 \
-initial-advertise-peer-urls http://$HOSTNAME:7001 -listen-peer-urls http://0.0.0.0:7001 \
-enable-v2"
echo "Running: $ETCD_CMD $*"
$ETCD_CMD "$@"
