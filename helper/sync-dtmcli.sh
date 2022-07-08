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
cp -rf ../dtm/dtmcli/* ./
rm -f *_test.go logger/*.log
sed -i '' -e 's/dtm-labs\/dtm\//dtm-labs\//g' *.go */**.go
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push
git tag $ver
git push --tags

cd ../dtmcli-go-sample
go get -u github.com/dtm-labs/dtmcli@$ver
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push


cd ../dtmgrpc
rm -rf *.go dtmgimp
cp -r ../dtm/dtmgrpc/* ./
go get github.com/dtm-labs/dtmcli@$ver
sed -i '' -e 's/dtm-labs\/dtm\//dtm-labs\//g' *.go */**.go
rm -rf *_test.go
rm -rf workflow/*_test.go
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push
git tag $ver
git push --tags

cd ../dtmgrpc-go-sample
go get github.com/dtm-labs/dtmcli@$ver
go get github.com/dtm-labs/dtmgrpc@$ver
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative busi/*.proto || exit 1
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push