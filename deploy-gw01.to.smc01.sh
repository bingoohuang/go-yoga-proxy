#!/usr/bin/env bash

deployName=$1
scp ./$deployName.linux.bin.bz2 smc01:app/$deployName/
ssh -tt smc01 "bash -s" << eeooff
cd app/$deployName
ps -ef|grep $deployName|grep -v grep|awk '{print \$2}'|xargs -r kill -9
rm $deployName.linux.bin
bzip2 -d $deployName.linux.bin.bz2
nohup ./$deployName.linux.bin -redisAddr=127.0.0.1:8051 > $deployName.out 2>&1 &
exit
eeooff