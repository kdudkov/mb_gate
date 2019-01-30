test:
	go test -v ./...
run:
	go run .
build:
	go build -o bin/mb_gate
	go build -o bin/client test_client/client.go
