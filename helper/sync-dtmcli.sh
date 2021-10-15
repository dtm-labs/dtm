#! /bin/bash
set -x
ver=$1
if [ x$ver == x ]; then
  echo please specify you version like vx.x.x;
  exit 1;
fi

if [ ${ver:0:1} != v ]; then
  echo please specify you version like vx.x.x;
  exit 1;
fi

cd ../dtmcli
cp ../dtm/dtmcli/*.go ./
rm -f *_test.go
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push
git tag $ver
git push --tags

cd ../dtmcli-go-sample
sleep 5
go get -u github.com/yedf/dtmcli
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push

cd ../dtmgrpc
cp ../dtm/dtmgrpc/*.go ./
cp ../dtm/dtmgrpc/*.proto ./
go get -u github.com/yedf/dtmcli
sed -i '' -e 's/yedf\/dtm\//yedf\//g' *.go *.proto
rm -rf *_test.go
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push
git tag $ver
git push --tags

cd ../dtmgrpc-go-sample
sleep 5
go get -u github.com/yedf/dtmcli
go get -u github.com/yedf/dtmgrpc
cp ../dtm/dtmgrpc/*.proto ./
sed -i '' -e 's/yedf\/dtm\//yedf\//g' *.go *.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative busi/*.proto || exit 1
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push