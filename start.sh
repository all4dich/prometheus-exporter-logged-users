#!/bin/bash
FULL_PATH=`realpath $0`
TARGET_DIR=`dirname $FULL_PATH`
PORT=49996
cd $TARGET_DIR
export PROCPS_USERLEN=20
export LC_ALL=C
export OS=$(uname | tr '[:upper:]' '[:lower:]')
export ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
  export ARCH="amd64"
fi
if [ "$ARCH" == "aarch64" ]; then
  export ARCH="arm64"
fi
cp prometheus-exporter-logged-users-$OS-$ARCH prometheus-exporter-logged-users
./prometheus-exporter-logged-users --port $PORT