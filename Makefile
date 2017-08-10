.DEFAULT_GOAL: ci
.PHONY: ci setup lint test

ci: setup test

setup:
	go get -t ./...
	go get github.com/alecthomas/gometalinter
	gometalinter --install

generate:
	go generate ./...

lint:
	gometalinter ./... --tests --deadline=5m --include=gofmt
	@echo lint passed

test:
	go test ./...
