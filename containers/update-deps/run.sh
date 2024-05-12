#!/bin/sh

log () { echo "$*" >&2; }

set -e

[ $# -eq 0 ] && set -- update-go update-js compact

cd /srv/heim/_deps

log "*** Merging split-apart files..."
node_modules/.bin/frankenstein recombine -v

for cmd; do
    log "*** Performing action $cmd..."
    ./deps.sh "$cmd" ..
done

log "*** Splitting large files..."
node_modules/.bin/frankenstein split -v

log

./deps.sh print-info ..

log
log OK
log
