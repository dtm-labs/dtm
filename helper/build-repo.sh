set -x
docker build -f helper/Dockerfile-release -t yedf/dtm:latest . && docker push yedf/dtm:latest