# !/bin/bash

cd /usr/share/sysbench/
echo 'create database sbtest;' > mysql -h 127.0.0.1 -uroot

sysbench oltp_write_only.lua --time=60 --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password= --mysql-db=sbtest --table-size=1000000 --tables=10 --threads=10 --events=999999999 --report-interval=10 prepare

sysbench oltp_write_only.lua --time=60 --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password= --mysql-db=sbtest --table-size=1000000 --tables=10 --threads=10 --events=999999999 --report-interval=10 run
