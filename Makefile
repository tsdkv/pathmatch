.PHONY: all test bench clean proto
test:
	go test -v ./...

bench:
	go test -bench=. -benchtime=5s ./...

clean:
	go clean -testcache

proto:
	protoc --go_out=. --go-grpc_out=. --proto_path=./proto ./proto/*.proto 

all: test bench
