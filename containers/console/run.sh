#!/bin/sh

# ssh really hates private key files with the wrong permissions and has no
# option to enhance its calm.
cp /keys/console_client* /keys/.copy/
chmod 0400 /keys/.copy/console_client

exec ssh -F /etc/heim/ssh_config "$@" backend
