#!/bin/bash

set -ex

go install \
    -ldflags "-X main.version=$(git --git-dir=src/euphoria.leet.nu/heim/.git rev-parse HEAD)" \
    euphoria.leet.nu/heim/heimctl

export PATH=/go/bin:"$PATH"
exec $*
