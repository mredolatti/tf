#!/usr/bin/env bash

HOST=${HOST:-'localhost'}
PORT=${PORT:-'5432'}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-'mysecretpassword'}
[[ ! -z "${POSTGRES_CONTAINER}" ]] && DOCKER_STR="docker exec -i ${POSTGRES_CONTAINER}"

export PGPASSWORD=$POSTGRES_PASSWORD;

function init_db() {

    local db=$1
    local db_user=$2
    local db_user_password=$3

    SCRIPT=$(cat <<EOSQL
        CREATE USER ${db_user} WITH PASSWORD '${db_user_password}';
        CREATE DATABASE ${db};
        GRANT ALL PRIVILEGES ON DATABASE ${db} TO ${db_user};
        \connect ${db} ${db_user}

        BEGIN;
            CREATE TABLE IF NOT EXISTS clients (
                id      VARCHAR NOT NULL PRIMARY KEY,
                secret  VARCHAR NOT NULL,
                domain  VARCHAR NOT NULL,
                user_id VARCHAR
            );
            CREATE TABLE IF NOT EXISTS tokens (
                client_id                       VARCHAR NOT NULL,
                user_id                         VARCHAR NOT NULL,
                redirect_uri                    VARCHAR NOT NULL,
                scope                           VARCHAR NOT NULL,
                code                            VARCHAR NOT NULL,
                code_created_at                 TIMESTAMPTZ,
                code_expires_in_seconds         INTEGER,
                code_challenge                  VARCHAR NOT NULL,
                code_challenge_method           VARCHAR NOT NULL,
                access                          VARCHAR NOT NULL,
                access_created_at               TIMESTAMPTZ,
                access_expires_in_seconds       INTEGER,
                refresh                         VARCHAR NOT NULL,
                refresh_created_at              TIMESTAMPTZ,
                refresh_expires_in_seconds      INTEGER
            );
        COMMIT;

        INSERT INTO clients(id, secret, domain, user_id)
        VALUES
            ('0000000', '1234567890', 'https://index-server:9876/api/clients/v1/accounts/auth_callback', '');
EOSQL
)

    echo "${SCRIPT}" | ${DOCKER_STR} psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER"
}

# TODO(mredolatti): is-client registration shold be done after initial self-register

init_db "filesrv"  "fs1" "${FS1_PASS}"
init_db "filesrv2" "fs2" "${FS2_PASS}"
