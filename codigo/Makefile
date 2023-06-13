.PHONY: all clean build test test-norace docker-images pki \
	runlocal-indexserver runlocal-fileserver fsbasic_for_debug incoming help

BASH ?= bash
GO ?= go
PROTOC ?= protoc
DOCKER ?= docker
DOCKER_COMPOSE ?= docker-compose
MAKE ?= make

sources := $(shell find . -name *.go)
grpc_generated := common/is2fs/changes.pb.go common/is2fs/changes_grpc.pb.go

# Por algun motivo la autogeneracion del target dejo de andar despues de un update a buildtools. debuggear algun dia
# pki_phony_tgts := $(shell cat PKI/Makefile | grep "\.PHONY" | sed 's/.PHONY: //')
# pki_all_tgts := $(shell make -f Makefile -C ./PKI -s list)
# pki_tgts := $(addprefix PKI/,$(filter-out $(pki_phony_tgts),$(pki_all_tgts)))
 
pki_tgts := PKI/root/private/ca.key \
			PKI/root/certs/ca.crt \
			PKI/sub/private/sub-ca.key \
			PKI/sub/certs/sub-ca.crt \
			PKI/indexserver/private/idx_server.key \
			PKI/indexserver/certs/idx_server.crt \
			PKI/indexserver/certs/chain.pem \
			PKI/fileserver/private/fs_server.key \
			PKI/fileserver/certs/fs_server.crt \
			PKI/fileserver/certs/chain.pem \
			PKI/fileserver2/private/fs_server.key \
			PKI/fileserver2/certs/fs_server.crt \
			PKI/fileserver2/certs/chain.pem \
			PKI/client/private/client.key \
			PKI/client/certs/client.crt

incoming_tgts := incoming/fileserver/ca.crt \
				 incoming/fileserver/chain.pem \
				 incoming/fileserver/fs_server.key \
				 incoming/fileserver2/ca.crt \
				 incoming/fileserver2/chain.pem \
				 incoming/fileserver2/fs_server.key \
				 incoming/indexserver/ca.crt \
				 incoming/indexserver/chain.pem \
				 incoming/indexserver/idx_server.key \
				 incoming/sqlinit/filesrvs.init.sh \
				 incoming/sqlinit/indexserver.init.sql 

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
	rm -f ./fsbasic.so

## Construir index-server, file-server y su respectivo plugin
build: index-server file-server fsbasic.so

## Construir index-server
index-server: $(sources) go.sum
	$(GO) build -o index-server ./indexsrv/cmd/server/main.go

## Construir file-server
file-server: $(sources) go.sum fsbasic.so 
	$(GO) build -o file-server ./fileserver/cmd/server/main.go

fsbasic.so: $(sources) go.sum
	$(GO) build -o fsbasic.so --buildmode=plugin ./fileserver/extension/plugins/fsbasic/plugin

fsbasic_for_debug: $(sources) go.sum
	$(GO) build -o fsbasic.so --buildmode=plugin -gcflags='all=-N -l' ./fileserver/extension/plugins/fsbasic/plugin

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
	MONGO_CONTAINER=mongo_mifs $(BASH) ./scripts/mongo_populator.sh

## Abrir un shell en mongodb
mongo-shell:
	$(DOCKER) exec -it infra-mongodb-1 mongosh --username root --password

# ------------ Targets para ejecucion local

## Ejecucion local de index-server con mongo como store
runlocal-indexserver: index-server $(pki_tgts)
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" \
		IS_DB_ENGINE="mongo" IS_MONGO_HOSTS="localhost:27017" IS_MONGO_DATABASE="mifs_indexsrv" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		./index-server

## Ejecucion local de index-server con postgres como store
runlocal-indexserver-postgres: index-server $(pki_tgts)
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" IS_PG_USER="postgres" \
		IS_PG_PWD="mysecretpassword" IS_PG_HOST="localhost" IS_PG_PORT="5432" \
		IS_PG_DB="indexsrv" IS_GOOGLE_CREDS_FN="./client_secret.json" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		./index-server

## Ejecucion local de file-server
runlocal-fileserver: file-server $(pki_tgts) fsbasic.so
	FS_LOG_DEBUG="true" FS_HOST="file-server-1" FS_PORT="9877" FS_SERVER_CERT_CHAIN="PKI/fileserver/certs/chain.pem" \
		FS_SERVER_PRIVATE_KEY="PKI/fileserver/private/fs_server.key" FS_ROOT_CA="PKI/root/certs/ca.crt" \
		FS_PSQL_URI="postgres://fs1:fs1@localhost:5432/filesrv" \
		FS_STORAGE_PLUGIN="./fsbasic.so" \
		FS_STORAGE_PLUGIN_CONF='{"filePath": "/home/martin/fsbasic/files", "authDBPath": "/home/martin/fsbasic/authdb"}' \
		./file-server


# ------------- Targets para debug

## Iniciar index-server en delve con mongodb como store
debug-indexserver:
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" \
		IS_DB_ENGINE="mongo" IS_MONGO_HOSTS="localhost:27017" IS_MONGO_DATABASE="mifs_indexsrv" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		dlv debug indexsrv/cmd/server/main.go

## Iniciar index-server en delve con postgres como store
debug-indexserver-postgres:
	IS_LOG_DEBUG="true" IS_HOST="index-server" IS_PORT="9876" IS_PG_USER="postgres" \
		IS_PG_PWD="mysecretpassword" IS_PG_HOST="localhost" IS_PG_PORT="5432" \
		IS_PG_DB="indexsrv" IS_GOOGLE_CREDS_FN="./client_secret.json" \
		IS_ROOT_CA="PKI/root/certs/ca.crt" IS_SERVER_CERT_CHAIN="PKI/indexserver/certs/chain.pem" \
		IS_SERVER_PRIVATE_KEY="PKI/indexserver/private/idx_server.key" \
		dlv debug indexsrv/cmd/server/main.go

## iniciar file-server en delve
debug-fileserver-postgres: fsbasic_for_debug
	FS_LOG_DEBUG="true" FS_HOST="file-server-1" FS_PORT="9877" FS_SERVER_CERT_CHAIN="PKI/fileserver/certs/chain.pem" \
		FS_SERVER_PRIVATE_KEY="PKI/fileserver/private/fs_server.key" FS_ROOT_CA="PKI/root/certs/ca.crt" \
		FS_PSQL_URI="postgres://fs1:fs1@localhost:5432/filesrv" \
		FS_STORAGE_PLUGIN="./fsbasic.so" \
		FS_STORAGE_PLUGIN_CONF='{"filePath": "/home/martin/fsbasic/files", "authDBPath": "/home/martin/fsbasic/authdb"}' \
		dlv debug fileserver/cmd/server/main.go

# ------------------------------------

## Run a CLI gRPC client against a local instance of file-server
runlocal-evans-fs: $(pki_tgts)
	evans \
		--host file-server \
		--port 9000 \
		--tls \
		--cacert PKI/root/certs/ca.crt \
		--proto protobuf/changes.proto

## Construir imagenes de docker para index-server y file-server
docker-images: infra/Dockerfile.index_server infra/Dockerfile.file_server
	$(DOCKER) build -t indexserver -f infra/Dockerfile.index_server .
	$(DOCKER) build -t fileserver -f infra/Dockerfile.file_server .

$(pki_tgts):
	$(MAKE) -f Makefile -C ./PKI all

docker-compose-build: infra/docker_compose.yml infra/index_server.Dockerfile infra/file_server.Dockerfile $(sources) go.sum $(incoming_tgts)
	$(DOCKER_COMPOSE) -f infra/docker_compose.yml build --no-cache 

docker-compose-up: docker-compose-build
	$(DOCKER_COMPOSE) -f infra/docker_compose.yml up

docker-compose-down: 
	$(DOCKER_COMPOSE) -f infra/docker_compose.yml down

## Preparar volumenes locales que seran montados en los contenedores de docker
incoming: $(incoming_tgts)

$(incoming_tgts): $(pki_tgts)
	mkdir -p incoming/{fileserver,fileserver2,indexserver}
	mkdir -p incoming/sqlinit
	cp PKI/root/certs/ca.crt incoming/fileserver
	cp PKI/fileserver/certs/chain.pem incoming/fileserver
	cp PKI/fileserver/private/fs_server.key incoming/fileserver
	cp PKI/root/certs/ca.crt incoming/fileserver2
	cp PKI/fileserver2/certs/chain.pem incoming/fileserver2
	cp PKI/fileserver2/private/fs_server.key incoming/fileserver2
	cp PKI/root/certs/ca.crt incoming/indexserver
	cp PKI/indexserver/certs/chain.pem incoming/indexserver
	cp PKI/indexserver/private/idx_server.key incoming/indexserver
	cp scripts/indexserver.init.sql incoming/sqlinit/
	cp scripts/filesrv.init.sh incoming/sqlinit/

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
