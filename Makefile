.DEFAULT_GOAL: ci
.PHONY: ci setup lint test

ci: setup test

setup:
	go get -t ./...
	gometalinter --install

generate:
	go generate ./...

lint:
	golangci-lint run
	@echo lint passed

test:
	go test ./...
