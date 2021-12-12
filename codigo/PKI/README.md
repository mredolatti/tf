Generar certificados:
=====================

1. Setear passphrases para los CA root e intermedio en variables de entorno:
`
export ROOT_CA_PASSPHRASE=somerelativelylongstring
export SUB_CA_PASSPHRASE=anotherrelativelylongstring
`

2. `make clean`
3. `make all`

Usar certificados:
==================

Agregar hosts alias al SO que mapeen los CN a localhost (o donde corran el index-server y file-server)
En linux, `/etc/hosts` =>
`
127.0.0.1 index-server
127.0.0.1 file-server
`
