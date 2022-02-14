#!/usr/bin/env bash

POSTGRES_HOST="index-server-db"
POSTGRES_PORT="5432"


waitport.sh ${POSTGRES_HOST} ${POSTGRES_PORT}

exec index-server

