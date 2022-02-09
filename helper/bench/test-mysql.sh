# !/bin/bash

set -x

cd /usr/share/sysbench/
echo 'create database sbtest;' > mysql -h 127.0.0.1 -uroot

sysbench oltp_write_only.lua --time=60 --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password= --mysql-db=sbtest --table-size=1000000 --tables=10 --threads=10 --events=999999999 --report-interval=10 prepare

sysbench oltp_write_only.lua --time=60 --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password= --mysql-db=sbtest --table-size=1000000 --tables=10 --threads=10 --events=999999999 --report-interval=10 run

export TIME=10
export CONCURRENT=20
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx&sqls=0" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx&sqls=5" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_barrier&sqls=5" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_tx&sqls=5" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx&sqls=1" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_barrier&sqls=1" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_tx&sqls=1" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_empty" && ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/bench"

