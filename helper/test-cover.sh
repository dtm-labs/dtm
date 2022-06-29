set -x
echo "mode: count" coverage.txt
for store in redis boltdb mysql postgres; do
  for d in $(go list ./... | grep -v vendor); do
    TEST_STORE=$store go test -failfast -covermode count -coverprofile=profile.out -coverpkg=github.com/dtm-labs/dtm/dtmcli,github.com/dtm-labs/dtm/dtmcli/dtmimp,github.com/dtm-labs/dtm/dtmcli/logger,github.com/dtm-labs/dtm/dtmgrpc,github.com/dtm-labs/dtm/dtmgrpc/workflow,github.com/dtm-labs/dtm/dtmgrpc/dtmgimp,github.com/dtm-labs/dtm/dtmsvr,github.com/dtm-labs/dtm/dtmsvr/config,github.com/dtm-labs/dtm/dtmsvr/storage,github.com/dtm-labs/dtm/dtmsvr/storage/boltdb,github.com/dtm-labs/dtm/dtmsvr/storage/redis,github.com/dtm-labs/dtm/dtmsvr/storage/registry,github.com/dtm-labs/dtm/dtmsvr/storage/sql,github.com/dtm-labs/dtm/dtmutil -gcflags=-l $d || exit 1
      if [ -f profile.out ]; then
          cat profile.out | grep -v 'mode:' >> coverage.txt
          echo > profile.out
      fi
  done
done

# go tool cover -html=coverage.txt

curl -s https://codecov.io/bash | bash
