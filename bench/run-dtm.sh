# !/bin/bash

# go run ../app/main.go
set -x
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

# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_empty" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=raw_tx" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_tx" && curl "http://127.0.0.1:8083/api/busi_bench/bench"
# curl "http://127.0.0.1:8083/api/busi_bench/reloadData?m=dtm_barrier" && curl "http://127.0.0.1:8083/api/busi_bench/bench"

ab -t $TIME -c $CONCURRENT "http://127.0.0.1:8083/api/busi_bench/benchEmptyUrl"