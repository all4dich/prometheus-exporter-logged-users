#!/bin/bash
FULL_PATH=`realpath $0`
TARGET_DIR=`dirname $FULL_PATH`
PORT=49996
cd $TARGET_DIR
./prometheus-exporter-logged-users --port $PORT