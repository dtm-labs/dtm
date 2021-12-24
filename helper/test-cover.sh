set -x
echo "" > coverage.txt
for store in redis mysql boltdb; do
  for d in $(go list ./... | grep -v vendor); do
    TEST_STORE=$store go test -covermode count -coverprofile=profile.out -coverpkg=github.com/yedf/dtm/common,github.com/yedf/dtm/dtmcli,github.com/yedf/dtm/dtmcli/dtmimp,github.com/yedf/dtm/dtmgrpc,github.com/yedf/dtm/dtmgrpc/dtmgimp,github.com/yedf/dtm/dtmsvr,github.com/yedf/dtm/dtmsvr/storage,github.com/yedf/dtm/dtmsvr/storage/boltdb,github.com/yedf/dtm/dtmsvr/storage/redis,github.com/yedf/dtm/dtmsvr/storage/registry,github.com/yedf/dtm/dtmsvr/storage/sql,github.com/yedf/dtm/dtmsvr/storage/boltdb,github.com/yedf/dtm/dtmsvr/storage/registry -gcflags=-l $d
      if [ -f profile.out ]; then
          cat profile.out >> coverage.txt
          echo > profile.out
      fi
  done
done

curl -s https://codecov.io/bash | bash
