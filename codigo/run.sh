#!/usr/bin/env bash

export GOOGLE_LOGIN_CLIENT_ID=$(cat client_secret.json | jq .web.client_id | sed -e s/'"'/''/g)
export GOOGLE_LOGIN_CLIENT_SECRET=$(cat client_secret.json | jq .web.client_secret | sed -e s/'"'/''/g)
./index-server
