# !/bin/bash
apt update
apt install -y git
git clone https://github.com/dtm-labs/dtm.git && cd dtm && git checkout alpha && cd bench && make


echo 'all prepared. you shoud run following commands to test in different terminal'
echo
echo 'cd dtm && go run bench/main.go redis|boltdb|db'
echo 'cd dtm && bench/run-redis|boltdb|mysql.sh'
