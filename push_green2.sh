#!/bin/bash
(cd carpe.pkg && make)
exec bash ../hodor/assimilate.sh carpe.pkg green2.yak.net
