#!/usr/bin/env bash

sleep 5

[[ -z "${MONGO_DB_INDEX}" ]] && MONGO_DB_INDEX='mifs_indexsrv'
[[ -z "${MONGO_DB_FILE}" ]] && MONGO_DB_FILE='mifs_filesrv'
[[ -z "${MONGO_HOST}" ]] && MONGO_HOST='localhost'

[[ ! -z "${MONGO_CONTAINER}" ]] && \
    DOCKER_STR="docker exec -i ${MONGO_CONTAINER}"

[[ ! -z "${MONGO_USER}" ]] && [[ ! -z "${MONGO_PASS}" ]] && \
    AUTH_STR="--username ${MONGO_USER} --password ${MONGO_PASS} --authenticationDatabase admin"

function mongo_query() {
    local db="${1}"
    local query="${2}"

    local auth_str=""
    local res=$(${DOCKER_STR} mongosh --host "${MONGO_HOST}" "${db}" ${AUTH_STR} --quiet --eval "${query}")
    echo "${res}"
}

function mongo_insert_one() {
    local db="${1}"
    local collection="${2}"
    local document="${3}"
    local res=$(mongo_query "${db}" "db.${collection}.insertOne(${document})")
    echo "${res}"
}

function f_org() {
    echo "{name: '${1}'}"
}

while getopts ":x" opt; do
    case $opt in
        l)
            listen_on_ready=1
            ;;
    esac
done


# Index-server application database index setup
mongo_query ${MONGO_DB_INDEX} "printjson(db.Organizations.createIndex({name: 1}, {unique: true}))"
mongo_query ${MONGO_DB_INDEX} "printjson(db.UserAccounts.createIndex({userId: 1}, {unique: false}))"
mongo_query ${MONGO_DB_INDEX} "printjson(db.UserAccounts.createIndex({userId: 1, organizationName: 1, serverName: 1}, {unique: true}))"
mongo_query ${MONGO_DB_INDEX} "printjson(db.FileServers.createIndex({organizationName: 1, name: 1}, {unique: true}))"
mongo_query ${MONGO_DB_INDEX} "printjson(db.Mappings.createIndex({userId: 1, path: 1}, {unique: true, partialFilterExpression: {path: {\$type: \"string\"}}}))"
mongo_query ${MONGO_DB_INDEX} "printjson(db.PendingOAuth2.createIndex({state: 1}, {unique: false}))"

# add one org
mongo_insert_one "${MONGO_DB_INDEX}" "Organizations" "$(f_org 'unicen')"

if [[ ! -z "${LISTEN_ON_READY}" ]]; then
    # accept TCP connection to allow other containers to wait for this one
    nc -l 0.0.0.0 5555
fi
