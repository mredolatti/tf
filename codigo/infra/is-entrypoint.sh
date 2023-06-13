#!/usr/bin/env bash

# TODO(mredolatti): usar nombrs genericos y recibir host t port por env
MONGO_HOST="mongodb"
MONGO_PORT="27017"

MDBP_HOST="mongodb_populator"
MDBP_PORT=5555

waitport.sh ${MONGO_HOST} ${MONGO_PORT}
waitport.sh ${MDBP_HOST} ${MDBP_PORT}

exec index-server
