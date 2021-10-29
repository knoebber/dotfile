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

dotfile:
	go build -o bin/dotfile cmd/dotfile/main.go

htmlgen:
	go build -o bin/htmlgen cmd/htmlgen/main.go

htmldocs: htmlgen
	bin/htmlgen -out server/html && bin/htmlgen -in docs/ -out server/html

dotfilehub: htmldocs
	go build -o bin/dotfilehub cmd/dotfilehub/main.go

clean:
	rm -rf bin/*

.PHONY: dotfile dotfilehub
