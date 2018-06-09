#!/bin/bash
(cd carpe.pkg && make)
exec bash ../hodor/assimilate.sh carpe.pkg aegon.yak.net
