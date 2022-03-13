.PHONY: all clean root sub fileserver indexserver client

folders := $(shell echo {root,sub,fileserver,indexserver,client}/{private,csr,crl,certs,newcerts})

$(folders):
	@mkdir -p $(folders)


## Armar todo el PKI
all: root sub fileserver indexserver client

## Armar clave privavada y certificado para el ROOT CA
root: root/private/ca.key root/certs/ca.crt

## Armar clave privavada y certificado para el CA intermedio (firmado por ROOT CA)
sub: sub/private/sub-ca.key sub/certs/sub-ca.crt

## Armar clave privavada y certificado para el file server (firmado por SUB CA)
fileserver: fileserver/private/fs_server.key fileserver/certs/chain.pem

## Armar clave privavada y certificado para el index server (firmado por SUB CA)
indexserver: indexserver/private/idx_server.key indexserver/certs/chain.pem

## Armar clave privavada y certificado para el index server (firmado por ROOT CA)
client: client/private/client.key client/certs/client.crt

# -------------------------------------
# Root CA
# -------------------------------------

root/private/ca.key: $(folders)
	touch root/index
	openssl rand -hex 16 > root/serial
	# generar clave privada encriptada con passphrase (leida desde env)
	openssl genrsa \
		-aes256 \
		-passout env:ROOT_CA_PASSPHRASE \
		-out root/private/ca.key 4096

root/certs/ca.crt: root/private/ca.key
	# generar certificado autofirmado
	PKI_CA_NAME="ca" PKI_CA_DIR="./root" openssl req -config openssl.conf \
		-key root/private/ca.key \
		-new \
		-x509 \
		-days 5000 \
		-sha256 \
		-extensions v3_ca \
		-passin env:ROOT_CA_PASSPHRASE \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=RootCA/emailAddress=some@localhost' \
		-out root/certs/ca.crt

# -------------------------------------
# Sub CA
# -------------------------------------

sub/private/sub-ca.key: $(folders)
	touch sub/index
	openssl rand -hex 16 > sub/serial
	# generar clave privada encriptada con passphrase (leida desde env)
	openssl genrsa \
		-aes256 \
		-passout env:SUB_CA_PASSPHRASE \
		-out sub/private/sub-ca.key 4096

sub/csr/sub-ca.csr: sub/private/sub-ca.key
	# Crear certificado + solicitud de firma
	PKI_CA_NAME="sub-ca" PKI_CA_DIR="./sub" openssl req -config openssl.conf \
		-new \
		-key sub/private/sub-ca.key \
		-sha256 \
		-passin env:SUB_CA_PASSPHRASE \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=SubCA/emailAddress=some@localhost' \
		-out sub/csr/sub-ca.csr

sub/certs/sub-ca.crt: sub/csr/sub-ca.csr root/private/ca.key root/certs/ca.crt
	# Firmar el certificado con el Root CA
	PKI_CA_NAME="ca" PKI_CA_DIR="./root" openssl ca -config openssl.conf \
		-batch \
		-extensions v3_intermediate_ca \
		-days 1000 \
		-notext \
		-passin env:ROOT_CA_PASSPHRASE \
		-in sub/csr/sub-ca.csr \
		-out sub/certs/sub-ca.crt

# -------------------------------------
# File server
# -------------------------------------

fileserver/private/fs_server.key: $(folders)
	# Generar clave privada (sin passphrase)
	openssl genrsa \
		-out fileserver/private/fs_server.key 2048

fileserver/csr/fs_server.csr: fileserver/private/fs_server.key
	# Crear certificado + solicitud de firma
	openssl req \
		-key fileserver/private/fs_server.key \
		-new \
		-sha256 \
		-addext 'subjectAltName = DNS:file-server,email:admin@file-server' \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=file-server/emailAddress=some@localhost' \
		-out fileserver/csr/fs_server.csr

fileserver/certs/fs_server.crt: fileserver/csr/fs_server.csr sub/private/sub-ca.key sub/certs/sub-ca.crt
	# Firmar certificado con el CA intermedio
	PKI_CA_NAME="sub-ca" PKI_CA_DIR="./sub" openssl ca -config openssl.conf \
		-batch \
		-extensions server_cert \
		-days 365 \
		-notext \
		-passin env:SUB_CA_PASSPHRASE \
		-in fileserver/csr/fs_server.csr \
		-out fileserver/certs/fs_server.crt

fileserver/certs/chain.pem: fileserver/certs/fs_server.crt sub/certs/sub-ca.crt
	# Armar certificate chain
	cat fileserver/certs/fs_server.crt sub/certs/sub-ca.crt > fileserver/certs/chain.pem

# -------------------------------------
# Index server
# -------------------------------------

indexserver/private/idx_server.key: $(folders)
	# Generar clave privada (sin passphrase)
	openssl genrsa \
		-out indexserver/private/idx_server.key 2048

indexserver/csr/idx_server.csr: indexserver/private/idx_server.key
	# Crear certificado + solicitud de firma
	openssl req \
		-key indexserver/private/idx_server.key \
		-new \
		-sha256 \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=index-server/emailAddress=some@localhost' \
		-addext 'subjectAltName = DNS:index-server' \
		-out indexserver/csr/idx_server.csr

indexserver/certs/idx_server.crt: indexserver/csr/idx_server.csr sub/private/sub-ca.key sub/certs/sub-ca.crt
	# Firmar certificado con el CA intermedio
	PKI_CA_NAME="sub-ca" PKI_CA_DIR="./sub" openssl ca -config openssl.conf \
		-batch \
		-extensions server_cert \
		-days 365 \
		-notext \
		-passin env:SUB_CA_PASSPHRASE \
		-in indexserver/csr/idx_server.csr \
		-out indexserver/certs/idx_server.crt


indexserver/certs/chain.pem: indexserver/certs/idx_server.crt sub/certs/sub-ca.crt
	# Armar certificate chain
	cat indexserver/certs/idx_server.crt sub/certs/sub-ca.crt > indexserver/certs/chain.pem

# -------------------------------------
# Client
# -------------------------------------

client/private/client.key: $(folders)
	# Generar claves y certificado p/cliente
	openssl genrsa \
		-out client/private/client.key 2048

client/csr/client.csr: client/private/client.key
	openssl req \
		-key client/private/client.key \
		-new \
		-sha256 \
		-addext 'subjectAltName = email:martinredolatti@gmail.com' \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=martin.redolatti/emailAddress=martinredolatti@gmail.com' \
		-out client/csr/client.csr

client/certs/client.crt: client/csr/client.csr root/private/ca.key
	# TODO: Usar una config especifica para clientes
	openssl x509 -req \
		-in client/csr/client.csr \
		-out client/certs/client.crt \
		-CA root/certs/ca.crt \
		-CAkey root/private/ca.key \
		-passin env:ROOT_CA_PASSPHRASE \
		-set_serial 01 -days 365

clean:
	rm -f root/{index,index.old,serial,serial.old,index.attr,index.attr.old}
	rm -f root/{private,certs,newcerts,crl,csr}/*
	rm -f sub/{index,index.old,serial,serial.old,index.attr,index.attr.old}
	rm -f sub/{private,certs,newcerts,crl,csr}/*
	rm -f fileserver/{private,certs,newcerts,crl,csr}/*
	rm -f indexserver/{private,certs,newcerts,crl,csr}/*
	rm -f client/{private,certs,newcerts,crl,csr}/*

all: root sub fileserver indexserver client

# tomado de: https://stackoverflow.com/a/26339924
list:
	@LC_ALL=C $(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null \
		| awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' \
		| sort \
		| egrep -v -e '^[^[:alnum:]]' -e '^$@$$'


