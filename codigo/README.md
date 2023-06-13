# MIFS

Las instrucciones a continuación explican los pasos para poder ejecutar el sistema en su totalidad de manera local.

## Armado del ambiente de pruebas

### Requerimientos
- GNU/Linux
- GNU/Make
- Docker
- Docker-Compose
- Compilador de C++ con soporte para C++20
- CMake
- Bibliotecas (SOs y sus correspondientes headers)
  - rapidjson
  - curl
  - openssl
  - libfuse
  - ImageMagick

### Creacion de infrastructura de claves publicas
- Exportar las variables de entorno que contengan las contraseñas a utilizar para los certificados raiz e intermedios:
```
export ROOT_CA_PASSPHRASE="some-root-pass"
export SUB_CA_PASSPHRASE="some-sub-pass"
```
- Invocar el target que genera y firma los certificados correspondientes
`make pki`

### Prepararacion de carpetas locales y lanzamiento de contenedores
- Invocar el target de armardo de carpetas compartidas con los contenedores (configuracion inicial de bases de datos, certificados, etc)
```
make incoming
```
- Levantar el entorno (servidores + bases de datos)
```
make docker-compose-up
```

### Construccion del driver y montaje de una vista haciendo uso del mismo
Las siguientes operaciones deben llevarse a cabo dentro de la sub-carpeta "driver".
- Construir driver (mount.mifs) y herramienta asociada (mifs-tools)
```
mkdir build
make cmake cmake-build
```
- Asegurarse que los hosts a referenciar sean accesibles por sus nombres, si se esta utilizando el ambiente de pruebas dockerizado, se debe actualizar el archivo `/etc/hosts` y agregar las siguientes lineas:
```
127.0.0.1 index-server
127.0.0.1 file-server-1
127.0.0.1 file-server-2
```
- Editar el archivo `config.json` y asegurarse que las direcciones de los servidores y los certificados de cliente asociados a cada uno son validos (en este ambiente de pruebas, el certificado de cliente es valido para ambos servidores).
- Crear una cuenta de usuario, iniciar sesion, configurar autenticacion en 2 pasos y volver a iniciar sesion
```
build/src/mifs-tools signup -c config.json -e "<direccion_email>" -u "<nombre y apellido>" -p "<contraseña>"
eval $(build/src/mifs-tools login -c config.json -e "<direccion_email>" -p "<contraseña>")
build/src/mifs-tools 2fa -c config.json

## Aca sele presentara una imagen que debera ser escaneada desde el telefono con una app tipo Google Authenticator o Authy

eval $(build/src/mifs-tools login -c config.json -e "<direccion_email>" -p "<contraseña>" -o <codigo_otp>)

```
- Vincular los servidores de archivos
```
build/src/mifs-tools link-server -c config.json -g unicen -s file-server-1
build/src/mifs-tools link-server -c config.json -g unicen -s file-server-2
```
- Montaje del sistema en una carpeta local
```
build/src/mount.mifs config.json <carpeta_local>
```
