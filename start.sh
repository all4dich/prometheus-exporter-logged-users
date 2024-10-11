#!/bin/bash
FULL_PATH=`realpath $0`
TARGET_DIR=`dirname $FULL_PATH`
PORT=49996
cd $TARGET_DIR
export PROCPS_USERLEN=20
export LC_ALL=C
./prometheus-exporter-logged-users --port $PORT