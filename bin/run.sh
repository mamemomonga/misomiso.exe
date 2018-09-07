#!/bin/bash
set -eu
BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
source $BASEDIR/build.env
cd $BASEDIR

rm -rf var/work
mkdir -p var/work

cp config.yaml var/work/

go build -o var/work/$APPNAME src/main.go 

cd var/work
exec ./$APPNAME
