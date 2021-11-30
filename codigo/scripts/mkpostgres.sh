#!/user/bin/env bash

# Initialize defaults if not set by user
[[ -z "${POSTGRES_CONTAINER}" ]] && POSTGRES_CONTAINER='some-postgres'
[[ -z "${POSTGRES_USER}" ]] && POSTGRES_USER='postgres'
[[ -z "${POSTGRES_PASS}" ]] && POSTGRES_PASS='mysecretpassword'
[[ -z "${POSTGRES_DB}" ]] && POSTGRES_DB='indexsrv'
[[ -z "${POSTGRES_HOST}" ]] && POSTGRES_HOST='localhost'

# Fetch the latest image if necessary, otherwhise do nothing.
# Return erroneous exit code if image fetch fails
function fetch_image() {
    images=$(docker images | grep postgres | wc -l)
    if [[ "$images" == "0" ]]; then
        docker pull postgres:latest
        return $?
    fi
}

# Create container if necessary. Otherwise try to start it
# If it exists and it's already running, no harm shold be done
function create_container() {
    existing=$(docker ps -a | grep postgres | awk '{print $1}')
    if [[ ! -z "$existing" ]]; then
        # container already exist, try to start it
        docker start $existing
        return $?
    fi
    docker run \
        --name "${POSTGRES_CONTAINER}" \
        -e POSTGRES_PASSWORD="${POSTGRES_PASS}" \
        -d \
        -p 5432:5432 \
        postgres
}

function drop_db() {
    uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}"    
    echo "DROP DATABASE ${POSTGRES_DB}" \
        | docker exec -i "${POSTGRES_CONTAINER}" psql ${uri}
}

function create_db() {
    uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}"    
    echo "CREATE DATABASE ${POSTGRES_DB} ENCODING 'utf-8'" \
        | docker exec -i "${POSTGRES_CONTAINER}" psql ${uri}
}

function init_db() {
    uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}/${POSTGRES_DB}"
    dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
    cat "${dir}/initdb.sql" | docker exec -i "${POSTGRES_CONTAINER}" psql "${uri}"
}

function setup_fixtures() {
    uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}/${POSTGRES_DB}"
    dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
    cat "${dir}/fixtures.sql" | docker exec -i "${POSTGRES_CONTAINER}" psql "${uri}"
}

function usage() {
      echo "$0 [-f]"
      echo " -f drop the database prior to attempting creation"
      exit 0
}

# ---- Main script execution ----

# Parse CLI args
while getopts ":hfd" opt; do
    case $opt in
        h)
            usage
            ;;
        d)
            drop_db=1
            ;;
        f)
            fixtures=1
    esac
done

fetch_image
create_container

if [[ ! -z "${drop_db}" ]]; then # Drop database if requested
    drop_db
fi

create_db
init_db

if [[ ! -z "${fixtures}" ]]; then # Insert fixtures/test data if requested
    setup_fixtures
fi
