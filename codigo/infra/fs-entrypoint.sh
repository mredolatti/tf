#!/usr/bin/env bash

# TODO(mredolatti): usar nombrs genericos y recibir host t port por env
PSQL_HOST="postgres"
PSQL_PORT="5432"

IS_HOST="index-server"
IS_PORT=9876

MDBP_HOST="mongodb_populator"
MDBP_PORT=5555

waitport.sh ${PSQL_HOST} ${PSQL_PORT}
waitport.sh ${IS_HOST} ${IS_PORT}

#waitport.sh ${MDBP_HOST} ${MDBP_PORT}

exec file-server
