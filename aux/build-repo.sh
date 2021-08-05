set -x
docker build -f aux/Dockerfile-release -t yedf/dtm:lastest . && docker push yedf/dtm:lastest