set -x
echo "mode: count" > coverage.txt
for store in redis boltdb mysql postgres; do
  for d in $(go list ./... | grep -v vendor); do
    TEST_STORE=$store go test -failfast -covermode count -coverprofile=profile.out -coverpkg=client/dtmcli,client/dtmcli/dtmimp,client/dtmcli/logger,client/dtmgrpc,client/workflow,client/dtmgrpc/dtmgimp,dtmsvr,dtmsvr/config,dtmsvr/storage,dtmsvr/storage/boltdb,dtmsvr/storage/redis,dtmsvr/storage/registry,dtmsvr/storage/sql,dtmutil -gcflags=-l $d || exit 1
      if [ -f profile.out ]; then
          cat profile.out | grep -v 'mode:' >> coverage.txt
          echo > profile.out
      fi
  done
done

# go tool cover -html=coverage.txt

curl -s https://codecov.io/bash | bash
