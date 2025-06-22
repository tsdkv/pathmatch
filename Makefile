.PHONY: all test bench clean proto
test:
	go test -v ./...

bench:
	go test -bench=. -benchtime=5s ./...

clean:
	go clean -testcache

proto:
	mkdir -p pathmatchpb
	protoc --proto_path=./proto --go_out=pathmatchpb --go_opt=paths=source_relative ./proto/pathmatch.proto

all: test bench
