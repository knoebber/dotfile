CLI_BINARY_NAME := dot
SERVER_BINARY_NAME := server

GO_TEST_FLAGS := -v -cover -count=1
GO_DEEP_TEST_FLAGS := $(GO_TEST_FLAGS) -race
CI_GO_TEST_FLAGS := $(GO_DEEP_TEST_FLAGS) -coverprofile=coverage.txt -covermode=atomic
GO_TEST_TARGET := ./...

test:
	go test $(GO_TEST_FLAGS) $(GO_TEST_TARGET)

deep_test:
	go test $(GO_DEEP_TEST_FLAGS) $(GO_TEST_TARGET)

ci_test:
	go test $(CI_GO_TEST_FLAGS) $(GO_TEST_TARGET)

cli:
	go build -o bin/$(CLI_BINARY_NAME) cli/*.go

server:
	go build -o bin/$(SERVER_BINARY_NAME) server/*.go

clean:
	rm -f bin/*

.PHONY: cli server
