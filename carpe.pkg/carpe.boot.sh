#!/bin/bash

set -x

mkdir -p /opt/carpe.spool
chown -R strick /opt/carpe.spool

rm ./spool
ln -s /opt/carpe.spool ./spool

su --command "./carpe-qrss-main >/tmp/carpe.log 2>&1 &" strick
