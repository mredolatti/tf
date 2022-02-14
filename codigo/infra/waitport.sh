#!/usr/bin/env bash

# HOST=$(echo "$1" | tr -d '\n\t ')
# PORT=$(echo "$2" | tr -d '\n\t ')

HOST=$1
PORT=$2

echo "Waiting for tcp://${HOST}:${PORT} to become ready"


function test_port() {
    nc -z "${HOST}" "${PORT}"
    return $?
}

test_port
while [ $? -ne 0 ]; do
    echo "Retrying..."
    sleep 1
    test_port
done

echo "tcp://${HOST}:${PORT} is ready!"
