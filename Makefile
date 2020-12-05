.PHONY: clean test build check fmt

default: clean check test build

clean:
	rm -f cover.out

build:
	go build

test:
	go test -v ./...

check:
	golangci-lint run
