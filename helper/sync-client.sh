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

cd ../client
cp -rf ../dtm/client/* ./
sed -i '' -e 's/dtm-labs\/dtm\//dtm-labs\//g' */*.go */*/*.go

rm -rf */*_test.go */*/*_test.go */*log */*/*log
go mod tidy
go build || exit 1

git add .
git commit -m"update from dtm to version $ver"
git push
git tag $ver
git push --tags

cd ../quick-start-sample

go get -u github.com/dtm-labs/client@$ver
go mod tidy
go build || exit 1
git add .
git commit -m"update from dtm to version $ver"
git push

