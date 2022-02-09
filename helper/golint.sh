set -x

curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.41.0
$(go env GOPATH)/bin/golangci-lint run
