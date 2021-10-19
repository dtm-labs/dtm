# !/bin/bash

# start all services
docker-compose -f helper/compose.mysql.yml up -d
go run app/main.go bench > /dev/nul