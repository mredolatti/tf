.PHONY: root sub indexserver fileserver

folders:
	@mkdir -p {root,sub,fileserver,indexserver,client}/{private,csr,crl,certs,newcerts}

root: folders
	touch root/index
	openssl rand -hex 16 > root/serial
	# generar clave privada encriptada con passphrase (leida desde env)
	openssl genrsa \
		-aes256 \
		-passout env:ROOT_CA_PASSPHRASE \
		-out root/private/ca.key 4096
	# generar certificado autofirmado
	openssl req -config root/root-ca.conf \
		-key root/private/ca.key \
		-new \
		-x509 \
		-days 5000 \
		-sha256 \
		-extensions v3_ca \
		-passin env:ROOT_CA_PASSPHRASE \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=RootCA/emailAddress=some@localhost' \
		-out root/certs/ca.crt

sub: folders root
	touch sub/index
	openssl rand -hex 16 > sub/serial
	# generar clave privada encriptada con passphrase (leida desde env)
	openssl genrsa \
		-aes256 \
		-passout env:SUB_CA_PASSPHRASE \
		-out sub/private/sub-ca.key 4096
	# Crear certificado + solicitud de firma
	openssl req -config sub/sub-ca.conf \
		-new \
		-key sub/private/sub-ca.key \
		-sha256 \
		-passin env:SUB_CA_PASSPHRASE \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=SubCA/emailAddress=some@localhost' \
		-out sub/csr/sub-ca.csr
	# Firmar el certificado con el Root CA
	openssl ca -config root/root-ca.conf \
		-batch \
		-extensions v3_intermediate_ca \
		-days 1000 \
		-notext \
		-passin env:ROOT_CA_PASSPHRASE \
		-in sub/csr/sub-ca.csr \
		-out sub/certs/sub-ca.crt

fileserver: folders sub
	# Generar clave privada (sin passphrase)
	openssl genrsa \
		-out fileserver/private/fs_server.key 2048
	# Crear certificado + solicitud de firma
	openssl req \
		-key fileserver/private/fs_server.key \
		-new \
		-sha256 \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=file-server/emailAddress=some@localhost' \
		-out fileserver/csr/fs_server.csr
	# Firmar certificado con el CA intermedio
	openssl ca -config sub/sub-ca.conf \
		-batch \
		-extensions server_cert \
		-days 365 \
		-notext \
		-passin env:SUB_CA_PASSPHRASE \
		-in fileserver/csr/fs_server.csr \
		-out fileserver/certs/fs_server.crt
	# Armar certificate chain
	cat fileserver/certs/fs_server.crt sub/certs/sub-ca.crt > fileserver/certs/chain.pem

indexserver: folders sub
	# Generar clave privada (sin passphrase)
	openssl genrsa \
		-out indexserver/private/idx_server.key 2048
	# Crear certificado + solicitud de firma
	openssl req \
		-key indexserver/private/idx_server.key \
		-new \
		-sha256 \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=index-server/emailAddress=some@localhost' \
		-out indexserver/csr/idx_server.csr
	# Firmar certificado con el CA intermedio
	openssl ca -config sub/sub-ca.conf \
		-batch \
		-extensions server_cert \
		-days 365 \
		-notext \
		-passin env:SUB_CA_PASSPHRASE \
		-in indexserver/csr/idx_server.csr \
		-out indexserver/certs/idx_server.crt
	# Armar certificate chain
	cat indexserver/certs/idx_server.crt sub/certs/sub-ca.crt > indexserver/certs/chain.pem

client: folders root
	# Generar claves y certificado p/cliente
	openssl genrsa \
		-out client/private/client.key 2048
	openssl req \
		-key client/private/client.key \
		-new \
		-sha256 \
		-subj '/C=AR/ST=Buenos Aires/L=Tandil/O=UNICEN/OU=IT/CN=martin.redolatti/emailAddress=martinredolatti@gmail.com' \
		-out client/csr/client.csr
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

all: clean root sub fileserver indexserver client
