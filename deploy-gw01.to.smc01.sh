#!/usr/bin/env bash

deployName=$1
scp ./$deployName.linux.bin smc01:.
ssh -tt smc01 "bash -s" << eeooff
cd app/$deployName
ps -ef|grep $deployName|grep -v grep|awk '{print \$2}'|xargs -r kill -9
mv -f ~/$deployName.linux.bin .
nohup ./$deployName.linux.bin -redisAddr=127.0.0.1:8051 > $deployName.out 2>&1 &
exit
eeooff