# All targets.
default: bench

bench: /usr/local/bin/go /etc/redis/redis.conf /usr/local/bin/docker-compose main.go
	rm -f ../conf.sample.yml
	go build -o bench

go: /usr/local/bin/go

redis: /etc/redis/redis.conf

mysql: /usr/local/bin/docker-compose

/usr/local/bin/go:
	wget https://golang.org/dl/go1.17.1.linux-amd64.tar.gz
	rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.1.linux-amd64.tar.gz && cp -f /usr/local/go/bin/go /usr/local/bin/go && rm go1.*

/etc/redis/redis.conf:
	apt update
	apt install -y redis redis-tools

/usr/local/bin/docker-compose:
	apt update
	apt install -y sysbench apache2-utils mysql-client-core-8.0
	curl -fsSL https://get.docker.com | sh
	curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
	chmod +x /usr/local/bin/docker-compose
	cd .. && docker-compose -f helper/compose.mysql.yml up -d && cd bench
