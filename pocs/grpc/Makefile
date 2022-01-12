GENERATED := user.pb.go user_grpc.pb.go
SOURCES := client/main.go server/main.go

all: $(GENERATED) $(SOURCES)
	go build -o client-app client/main.go
	go build -o server-app server/main.go

$(GENERATED) : user.proto
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative user.proto
