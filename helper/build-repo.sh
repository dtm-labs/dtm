set -x
docker build -f helper/Dockerfile-release -t yedf/dtm:lastest . && docker push yedf/dtm:lastest