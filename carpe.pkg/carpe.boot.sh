#!/bin/bash

set -x

mkdir -p /opt/carpe.spool
chown -R strick /opt/carpe.spool

rm ./spool
ln -s /opt/carpe.spool ./spool

su --command "
  while date;
  do
    find /opt/carpe.spool/ -type f -mtime +4 -print | xargs  rm -f
    ./surveyutil >/tmp/survey.log 2>&1
    sleep 300
  done &" strick
su --command "./carpe-qrss-main >/tmp/carpe.log 2>&1 &" strick
