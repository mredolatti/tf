#!/usr/bin/env bash

email=$1
if [ -z $email ]; then
    echo "usage: bash demo.sh <email-address>"
    exit 1
fi

build/src/mifs-tools signup -c config.json -e ${email} -u martin -p 123456

token_cmd=$(build/src/mifs-tools login -c config.json -e ${email} -p 123456)
echo "[debug] - ${token_cmd}"
eval ${token_cmd}

build/src/mifs-tools 2fa -c config.json

echo -e "\n\nInsert OTP code: "
read otp

token_cmd=$(build/src/mifs-tools login -c config.json -e ${email} -p 123456 -o ${otp})
echo "[debug] - ${token_cmd}"
eval ${token_cmd}

build/src/mifs-tools link-server -c config.json -g unicen -s file-server-1
build/src/mifs-tools link-server -c config.json -g unicen -s file-server-2
build/src/mount.mifs config.json prueba/
