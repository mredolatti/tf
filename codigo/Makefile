.PHONY: all clean build build_indexserver test testrace

all: test testrace build

clean:
	rm -f ./index-server
	rm -f ./file-server

build: build_indexserver build_fileserver

build_indexserver:
	@go build -o index-server ./indexsrv/cmd/server/main.go

build_fileserver:
	@go build -o file-server ./fileserver/cmd/server/main.go

test:
	@go test ./... -count=1

testrace:
	@go test ./... -count=1 -race
