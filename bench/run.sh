# !/bin/bash

# go run ../app/main.go
set -x
TIME=1
CURRENT=10
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_empty" && ab -t $TIME -c $CURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_tx" && ab -t $TIME -c $CURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_empty" && ab -t $TIME -c $CURRENT "http://127.0.0.1:8083/api/busi_bench/bench"
curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx" && ab -t $TIME -c $CURRENT "http://127.0.0.1:8083/api/busi_bench/bench"

# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_empty" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_tx" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_empty" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
