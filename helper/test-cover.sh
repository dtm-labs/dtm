set -x
export DTM_DEBUG=1
echo "mode: count" > coverage.txt
for store in redis boltdb mysql postgres; do
  TEST_STORE=$store go test -failfast -covermode count -coverprofile=profile.out -coverpkg=github.com/dtm-labs/dtm/client/dtmcli,github.com/dtm-labs/dtm/client/dtmcli/dtmimp,github.com/dtm-labs/logger,github.com/dtm-labs/dtm/client/dtmgrpc,github.com/dtm-labs/dtm/client/workflow,github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp,github.com/dtm-labs/dtm/dtmsvr,dtmsvr/config,github.com/dtm-labs/dtm/dtmsvr/storage,github.com/dtm-labs/dtm/dtmsvr/storage/boltdb,github.com/dtm-labs/dtm/dtmsvr/storage/redis,github.com/dtm-labs/dtm/dtmsvr/storage/registry,github.com/dtm-labs/dtm/dtmsvr/storage/sql,github.com/dtm-labs/dtm/dtmutil -gcflags=-l ./... || exit 1
    echo "TEST_STORE=$store finished"
    if [ -f profile.out ]; then
        cat profile.out | grep -v 'mode:' >> coverage.txt
        echo > profile.out
    fi
done
## for local unit test, you may use following command
# SKIP_MONGO=1 TEST_STORE=redis GOARCH=amd64 go test -v -failfast -count=1  -gcflags=all=-l ./...

# go tool cover -html=coverage.txt

# curl -s https://codecov.io/bash | bash
