#! /bin/bash
set -x
if [ x$1 == x ]; then
  echo please specify you version like vx.x.x;
  exit 1;
fi

if [ ${1:1:1} != v ]; then
  echo please specify you version like vx.x.x;
  exit 1;
fi

cd ../dtmcli
cp ../dtm/dtmcli/*.go ./
rm -f *_test.go
go mod tidy
go build || exit 1

git add .
git commit -m'update from dtm'
git push
# git tag $1
# git push --tags

cd ../dtmcli-go-sample
go get -u github.com/yedf/dtmcli
go mod tidy
go build || exit 1
git add .
git commit -m'update from dtm'
git push

cd ../dtmgrpc
cp ../dtm/dtmgrpc/*.go ./
cp ../dtm/dtmgrpc/*.proto ./

sed -i '' -e 's/yedf\/dtm\//yedf\//g' *.go *.proto
rm -rf *_test.go
go get -u github.com/yedf/dtmcli
go mod tidy
go build || exit 1
git add .
git commit -m'update from dtm'
git push
# git tag $1
# git push --tags

cd ../dtmgrpc-go-sample
go get -u github.com/yedf/dtmcli
go get -u github.com/yedf/dtmgrpc
go build || exit 1
git add .
git commit -m'update from dtm'
git push