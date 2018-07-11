#!/usr/bin/env bash

# ./deploy.sh yogaapp@test.ino01
targetHost=$1
deployName=go-yoga-proxy
fast=$2

if [ "$fast" == "fast" ]; then
    echo "jump building in fast mode"
else
    echo "rebuilding"
    env GOOS=linux GOARCH=amd64 go build -o $deployName.linux.bin
    upx $deployName.linux.bin
fi

rsync -avz --human-readable --progress -e "ssh -p 22" ./$deployName.linux.bin $targetHost:.
ssh -tt $targetHost "bash -s" << eeooff
cd ./app/$deployName/
ps -ef|grep $deployName|grep -v grep|awk '{print \$2}'|xargs -r kill -9
mv -f ~/$deployName.linux.bin .
nohup ./$deployName.linux.bin -redisAddr=127.0.0.1:8051 > $deployName.out 2>&1 &
exit
eeooff
