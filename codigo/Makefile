.PHONY: all clean build build_indexserver test testrace

all: test testrace build

clean:
	@[[ -f ./index-server ]] && rm ./index-server

build: build_indexserver

build_indexserver:
	@go build -o index-server ./indexsrv/cmd/server/main.go

test:
	@go test ./... -count=1

testrace:
	@go test ./... -count=1 -race
