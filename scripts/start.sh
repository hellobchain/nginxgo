#!/bin/bash
NGINXGO_BIN="./bin/nginxgo.bin"
nohup ./${NGINXGO_BIN} start >nginxgo.log 2>&1 &