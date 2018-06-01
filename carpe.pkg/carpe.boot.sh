#!/bin/bash

mkdir -p /opt/carpe.spool
chown strick /opt/carpe.spool
su --command "./carpe-qrss --delay=4m --spool=/opt/carpe.spool/ >/tmp/carpe.log 2>&1 &" strick
