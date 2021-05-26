default: tests

tests: golangci-lint unittest

unittest:
	go test ./...

golangci-lint:
	golangci-lint run ./...

.PHONY: tests unittest golangci-lint
