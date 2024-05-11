#!/bin/sh

set -e

cd /srv/heim/_deps

node_modules/.bin/frankenstein recombine -v

./deps.sh update-go ..
./deps.sh update-js ..
./deps.sh compact ..

node_modules/.bin/frankenstein split -v

echo

./deps.sh print-info ..

echo
echo OK
echo
