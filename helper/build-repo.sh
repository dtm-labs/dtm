set -x
if [ x$1 == x ]; then
  echo please specify version
  exit 1
fi

docker build -f helper/Dockerfile-release -t yedf/dtm:latest . && docker push yedf/dtm:latest
docker tag yedf/dtm:latest yedf/dtm:$1 && docker push yedf/dtm:$1