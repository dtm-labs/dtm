# !/bin/bash
apt update
apt install -y software-properties-common
add-apt-repository -y ppa:redislabs/redis
apt install -y redis redis-tools

