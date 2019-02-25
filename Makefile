.PHONY: clean build

default: all

all: test build

GIT_REVISION=`git rev-parse --short HEAD`
GIT_BRANCH=`git rev-parse --symbolic-full-name --abbrev-ref HEAD`

LDFLAGS=-ldflags "-s -X main.gitRevision=${GIT_REVISION} -X main.gitBranch=${GIT_BRANCH}"

clean:
	rm bin/*
test:
	go test -v ./...
run:
	go run .
build:
	go mod tidy
	go build ${LDFLAGS} -o bin/mb_gate
	go build ${LDFLAGS} -o bin/client test_client/client.go
