BINARY_NAME := dot

GO_TEST_FLAGS := -v -cover -count=1
GO_DEEP_TEST_FLAGS := $(GO_TEST_FLAGS) -race
GO_TEST_TARGET := ./commands/...

test:
	go test $(GO_TEST_FLAGS) $(GO_TEST_TARGET)

deep_test:
	go test $(GO_DEEP_TEST_FLAGS) $(GO_TEST_TARGET)

ci_test: deep_test

binary:
	env GO111MODULE=on go build -o bin/$(BINARY_NAME) .

clean:
	rm -f bin/*

