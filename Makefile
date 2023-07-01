# dev env https://www.dtm.pub/other/develop.html
all: fmt lint test_redis
.PHONY: all

fmt:
	@gofmt -s -w ./

lint:
	revive -config revive.toml ./...

.PHONY: test
test:
	@go test ./...

test_redis:
	TEST_STORE=redis go test ./...

test_all:
	TEST_STORE=redis go test ./...
	TEST_STORE=boltdb go test ./...
	TEST_STORE=mysql go test ./...
	TEST_STORE=postgres go test ./...

cover_test:
	./helper/test-cover.sh

