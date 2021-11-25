all: test testrace build

build: build_testserver

build_testserver:
	go build -o index-server ./indexsrv/cmd/server/main.go

test:
	go test ./... -count=1

testrace:
	go test ./... -count=1 -race
