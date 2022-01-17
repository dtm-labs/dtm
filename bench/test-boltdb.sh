# !/bin/bash

set -x

ab -n 50000 -c 10 "http://127.0.0.1:8083/api/busi_bench/benchEmptyUrl"
