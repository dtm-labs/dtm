# !/bin/bash

set -x

export LOG_LEVEL=fatal
export STORE_DRIVER=redis
export STORE_HOST=localhost
export STORE_PORT=6379
go build
./bench redis &
echo 'sleeping 3s for dtm bench to run up.' && sleep 3
curl "http://127.0.0.1:8083/api/busi_bench/benchFlashSalesReset"
ab -n 1000 -c 20 "http://127.0.0.1:8083/api/busi_bench/benchFlashSales"
pkill bench
