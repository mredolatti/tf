#!/usr/bin/env bash

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
    local OPTIND
    while getopts "i:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z ${fid+x} ] && echo ${usage} && return 1

    curl \
        -XGET \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/files/${fid}"
}

function fs_create() {
    local usage="usage: fs_create -n <name> -s <notes> -p <patient> -t <type>"
    local OPTIND
    while getopts "hn:s:p:t:" option; do
        case ${option} in
            n) local name=${OPTARG} ;;
            s) local notes=${OPTARG} ;;
            p) local patient=${OPTARG} ;;
            t) local typ=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${name+x}" ] || [ -z "${notes+x}" ] || [ -z "${patient+x}" ] || [ -z "${typ+x}" ] \
        && echo ${usage} && return 1

    curl \
        -XPOST \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        -H'Content-Type: application/json' \
        -d"{\"name\": \"${name}\", \"notes\": \"${notes}\", \"patientId\": \"${patient}\", \"type\": \"${typ}\"}" \
        'https://file-server:9877/files'
}

function fs_update() {
    local usage="usage: fs_update -i <id> -n <name> -s <notes> -p <patient> -t <type>"
    local OPTIND
    while getopts "hi:n:s:p:t:" option; do
        case ${option} in
            i) local id=${OPTARG} ;;
            n) local name=${OPTARG} ;;
            s) local notes=${OPTARG} ;;
            p) local patient=${OPTARG} ;;
            t) local typ=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${name+x}" || -z "${notes+x}" ||-z "${patient+x}" ||-z "${typ+x}" ] \
        && echo ${usage} && return 1

    curl \
        -XPUT \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        -H'Content-Type: application/json' \
        -d"{\"name\": \"${name}\", \"notes\": \"${notes}\", \"patientId\": \"${patient}\", \"type\": \"${typ}\"}" \
        "https://file-server:9877/files/${id}"
}

function fs_del() {
    usage="usage: fs_del -i <file_id>"
    local OPTIND
    while getopts "i:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${fid+x}" ] && echo ${usage} && return 1

    curl \
        -XDELETE \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/files/${fid}"
}

function fs_get_contents() {
    usage="usage: fs_get_contents -i <file_id>"
    local OPTIND
    while getopts "i:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${fid+x}" ] && echo ${usage} && return 1

    curl \
        -XGET \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/files/${fid}/contents"
}

function fs_update_contents() {
    usage="usage: fs_update_contents -i <file_id>"
    local OPTIND
    while getopts "i:f:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            f) local fname=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${fid+x}" ] || [ -z "${fname+x}" ] && echo ${usage} && return 1

    curl \
        -XPUT \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        --data-binary "@${fname}" \
        "https://file-server:9877/files/${fid}/contents"
}

function fs_delete_contents() {
    usage="usage: fs_delete_contents -i <file_id>"
    local OPTIND
    while getopts "i:" options; do
        case ${options} in
            i) local fid=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${fid+x}" ] && echo ${usage} && return 1

    curl \
        -XDELETE \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/files/${fid}/contents"
}
