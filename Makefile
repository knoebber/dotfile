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

dot:
	go build -o bin/dot cmd/dot/*.go

dotfilehub: clean
	go build -o bin/dotfilehub cmd/dotfilehub/*.go && \
		cp -r cmd/dotfilehub/assets bin/assets && \
                cp -r cmd/dotfilehub/tmpl bin/tmpl

clean:
	rm -rf bin/*

.PHONY: cli dotfilehub
