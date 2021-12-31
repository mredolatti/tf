#!/bin/bash

FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
FS_CERT=${FS_CERT:-'PKI/client/certs/client.crt'}
FS_KEY=${FS_KEY:-'PKI/client/private/client.key'}

function fs_list() {
    curl \
        -XGET \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        'https://file-server:9877/files'
}

function fs_get() {
    usage="usage: fs_get -i <file_id>"
    while getopts "i:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            :) echo ${usage} ;;
            *) echo ${usage} ;;
        esac
    done

    curl \
        -XGET \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/files/${fid}"
}

function fs_create() {
    usage="usage: fs_create -n <name> -s <notes> -p <patient> -t <type>"
    while getopts "n:s:p:t:" options; do
        case ${options} in
            n) local name=${OPTARG} ;;
            s) local notes=${OPTARG} ;;
            p) local patient=${OPTARG} ;;
            t) local typ=${OPTARG} ;;
            :) echo ${usage} ;;
            *) echo ${usage} ;;
        esac
    done

    curl \
        -XPOST \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        -H'Content-Type: application/json' \
        -d"{\"name\": \"${name}\", \"notes\": \"${notes}\", \"patientId\": \"${patient}\", \"type\": \"${typ}\"}" \
        'https://file-server:9877/files'
}



