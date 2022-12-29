.PHONY: all clean build test test-norace docker-images pki \
	runlocal-indexserver runlocal-fileserver help

BASH ?= bash
GO ?= go
PROTOC ?= protoc
DOCKER ?= docker
MAKE ?= make

sources := $(shell find . -name *.go)
grpc_generated := common/is2fs/changes.pb.go common/is2fs/changes_grpc.pb.go
pki_phony_tgts := $(shell cat PKI/Makefile | grep "\.PHONY" | sed 's/.PHONY: //')
pki_all_tgts := $(shell make -f Makefile -C ./PKI -s list)
pki_tgts := $(addprefix PKI/,$(filter-out $(pki_phony_tgts),$(pki_all_tgts)))

container_incoming_tgts := incoming/fileserver/certs/ca.crt \
			   incoming/fileserver/certs/key.pem \
			   incoming/fileserver/certs/chain.pem

default: help

## Armar infra de claves publicas, compilar y correr tests
all: pki build test

$(grpc_generated): protobuf/changes.proto
	$(PROTOC) \
		--go_out=./common/is2fs \
		--go_opt=paths=source_relative \
		--go-grpc_out=./common/is2fs \
		--go-grpc_opt=paths=source_relative \
		--proto_path=./protobuf \
		changes.proto

## Actualizar dependencias
go.sum: go.mod $(grpc_generated) # generated files are added for proper 3rd party deps tracking
	$(GO) mod tidy

## Eliminar archivos generados (certificados, binarios, etc)
clean:
	$(MAKE) -f Makefile -C ./PKI clean
	rm -Rf ./incoming
	rm -f ./index-server
	rm -f ./file-server

## Construir index-server y file-server
build: index-server file-server

## Construir index-server
index-server: $(sources) go.sum
	$(GO) build -o index-server ./indexsrv/cmd/server/main.go

## Construir file-server
file-server: $(sources) go.sum
	$(GO) build -o file-server ./fileserver/cmd/server/main.go

## Ejecutar tests incluyando analisis de race conditions
test:
	$(GO) test ./... -count=1 -race

## Ejecutar tests sin analisis de race conditions
test-norace:
	$(GO) test ./... -count=1

## Armar infra de claves publicas
pki: $(pki_tgts)

## Configurar postgres local desde un contenedor
psql-rebuild:
	$(BASH) ./scripts/mkpostgres.sh -dcf

## Abrir un shell en postgres
psql-shell:
	$(BASH) ./scripts/mkpostgres.sh -s

## Configurar mongodb local desde un contenedor
mongo-rebuild:
	$(BASH) ./scripts/mkmongo.sh -dcf

## Abrir un shell en mongodb
mongo-shell:
	$(BASH) ./scripts/mkmongo.sh -s

## Ejecucion local de index-server
runlocal-indexserver: index-server $(pki_tgts)
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" \
		IS_DB_ENGINE="mongo" IS_MONGO_HOSTS="localhost:27017" \
		IS_MONGO_DATABASE="mifs_indexsrv" IS_GOOGLE_CREDS_FN="./client_secret.json" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		./index-server

## Ejecucion local de index-server
runlocal-indexserver-postgres: index-server $(pki_tgts)
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" IS_PG_USER="postgres" \
		IS_PG_PWD="mysecretpassword" IS_PG_HOST="localhost" IS_PG_PORT="5432" \
		IS_PG_DB="indexsrv" IS_GOOGLE_CREDS_FN="./client_secret.json" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		./index-server

## Iniciar index-server en delve
debuglocal-indexserver-postgres:
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" IS_PG_USER="postgres" \
		IS_PG_PWD="mysecretpassword" IS_PG_HOST="localhost" IS_PG_PORT="5432" \
		IS_PG_DB="indexsrv" IS_GOOGLE_CREDS_FN="./client_secret.json" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		dlv debug indexsrv/cmd/server/main.go

## Ejecucion local de file-server
runlocal-fileserver-postgres: file-server $(pki_tgts)
	FS_LOG_DEBUG="true" FS_HOST="file-server" FS_PORT="9877" FS_SERVER_CERT_CHAIN="PKI/fileserver/certs/chain.pem" \
		FS_PG_USER="postgres" FS_PG_PWD="mysecretpassword" FS_PG_HOST="localhost" FS_PG_PORT="5432" FS_PG_DB="filesrv" \
		FS_SERVER_PRIVATE_KEY="PKI/fileserver/private/fs_server.key" FS_ROOT_CA="PKI/root/certs/ca.crt" \
		./file-server

## iniciar file-server en delve
debuglocal-fileserver-postgres:
	FS_LOG_DEBUG="true" FS_HOST="file-server" FS_PORT="9877" FS_SERVER_CERT_CHAIN="PKI/fileserver/certs/chain.pem" \
		FS_PG_USER="postgres" FS_PG_PWD="mysecretpassword" FS_PG_HOST="localhost" FS_PG_PORT="5432" FS_PG_DB="filesrv" \
		FS_SERVER_PRIVATE_KEY="PKI/fileserver/private/fs_server.key" FS_ROOT_CA="PKI/root/certs/ca.crt" \
		dlv debug fileserver/cmd/server/main.go


## Run a CLI gRPC client against a local instance of file-server
runlocal-evans-fs:
	evans --host localhost --port 9000 --proto protobuf/changes.proto

## Construir imagenes de docker para index-server y file-server
docker-images: infra/Dockerfile.index_server infra/Dockerfile.file_server
	$(DOCKER) build -t indexserver -f infra/Dockerfile.index_server .
	$(DOCKER) build -t fileserver -f infra/Dockerfile.file_server .

$(pki_tgts):
	$(MAKE) -f Makefile -C ./PKI all

docker-compose-build: infra/Dockerfile.index_server infra/Dockerfile.file_server $(sources) go.sum $(container_incoming_tgts)
	docker-compose -f infra/docker_compose.yml build

docker-compose-up: docker-compose-build
	docker-compose -f infra/docker_compose.yml up

docker-compose-down: 
	docker-compose -f infra/docker_compose.yml down

## Preparar volumenes locales que seran montados en los contenedores de docker
incoming: $(container_incoming_tgts)

$(container_incoming_tgts): $(pki_tgts)
	mkdir -p incoming/{fileserver,indexserver}/certs
	mkdir -p incoming/indexserver/credentials
	mkdir -p incoming/idx_server_db/scripts
	cp PKI/root/certs/ca.crt incoming/fileserver/certs
	cp PKI/fileserver/certs/chain.pem incoming/fileserver/certs
	cp PKI/fileserver/private/fs_server.key incoming/fileserver/certs
	cp PKI/root/certs/ca.crt incoming/indexserver/certs
	cp PKI/indexserver/certs/chain.pem incoming/indexserver/certs
	cp PKI/indexserver/private/idx_server.key incoming/indexserver/certs
	cp client_secret.json incoming/indexserver/credentials
	cp scripts/fileserver.init.sql incoming/idx_server_db/scripts
	cp scripts/indexserver.init.sql incoming/idx_server_db/scripts


# autodoc de targets tomado de: https://docs.cloudposse.com/reference/best-practices/make-best-practices/
## Esta ayuda
help:
	@printf "Targets disponibles:\n\n"
	@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
	    helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
		    helpCommand = $$1; \
		    helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
		    gsub("\\\\", "", helpCommand); \
		    gsub(":+$$", "", helpCommand); \
		    printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	    } \
	    { lastLine = $$0 }' $(MAKEFILE_LIST) 2> /dev/null | sort -u
	@printf "\n"
