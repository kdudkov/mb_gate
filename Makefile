all: build

test:
	go test -v ./...
run:
	go run .
build:
	go mod tidy
	go build -o bin/mb_gate
	go build -o bin/client test_client/client.go
