BINARY_NAME := dot

GO_TEST_FLAGS := -v -cover -count=1
GO_DEEP_TEST_FLAGS := $(GO_TEST_FLAGS) -race
CI_GO_TEST_FLAGS := $(GO_DEEP_TEST_FLAGS) -coverprofile=coverage.txt -covermode=atomic
GO_TEST_TARGET := ./commands/...

test:
	go test $(GO_TEST_FLAGS) $(GO_TEST_TARGET)

deep_test:
	go test $(GO_DEEP_TEST_FLAGS) $(GO_TEST_TARGET)

ci_test:
	go test $(CI_GO_TEST_FLAGS) $(GO_TEST_TARGET)

binary:
	go build -o bin/$(BINARY_NAME) .

clean:
	rm -f bin/*

