# !/bin/bash

# install all commands needed

apt update
apt install -y sysbench apache2-utils mysql-client-core-8.0 redis redis-tools

# install docker and docker-compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# install go
wget https://golang.org/dl/go1.17.1.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.1.linux-amd64.tar.gz && cp -f /usr/local/go/bin/go /usr/local/bin/go
