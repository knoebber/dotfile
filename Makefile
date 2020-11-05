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

bin/assets:
	cp -r server/assets bin/assets

bin/templates:
	cp -r server/templates bin/templates

bin/html:
	cp -r server/html bin/html

htmldocs: htmlgen bin/html
	bin/htmlgen -out bin/html && bin/htmlgen -in docs/ -out bin/html

dotfilehub: clean bin/templates bin/assets htmldocs
	go build -o bin/dotfilehub cmd/dotfilehub/main.go

clean:
	rm -rf bin/*

.PHONY: dotfile dotfilehub
