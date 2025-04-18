#!/bin/bash

NGINXGO_BIN="./bin/nginxgo.bin"
pid=`ps -ef | grep ${NGINXGO_BIN} | grep -v grep | awk '{print $2}'`
if [ ! -z ${pid} ];then
    echo "kill $pid"
    kill $pid
else
    echo "$NGINXGO_BIN already stopped"
fi