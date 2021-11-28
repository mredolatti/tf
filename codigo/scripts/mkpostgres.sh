#!/user/bin/env bash

# Fetch the latest image if necessary, otherwhise do nothing.
# Return erroneous exit code if image fetch fails
function get_image() {
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
    [[ -z "${POSTGRES_CONTAINER}" ]] && postgres_container='some-postgres' || postgres_container="${POSTGRES_CONTAINER}"
    [[ -z "${POSTGRES_PASS}" ]] && postgress_pass='mysecretpassword' || postgress_pass="${POSTGRESS_PASS}"
    docker run \
        --name "${POSTGRES_CONTAINER}" \
        -e POSTGRES_PASSWORD="${POSTGRES_PASS}" \
        -d \
        -p 5432:5432 \
        postgres
}


get_image && create_container
