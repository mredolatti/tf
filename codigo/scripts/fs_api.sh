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
    usage="usage: fs_update_contents -i <file_id> -f <path-to-file>"
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

function fs_auth_code() {

    usage="usage: fs_auth_code -i <client_id>"
    local OPTIND
    while getopts "i:" options; do
        case ${options} in
            i) local cid=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z "${cid+x}" ] && echo ${usage} && return 1

    curl \
        -XGET -v \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "https://file-server:9877/authorize?client_id=${cid}&response_type=code"
}

function fs_exchange_code() {

    usage="usage: fs_exchange_code -c <code> -i <client_id> -s <client_secret> -r <redirect_uri>"
    local OPTIND
    while getopts "c:i:s:r:" options; do
        case ${options} in
            c) local code=${OPTARG} ;;
            i) local cid=${OPTARG} ;;
            s) local secret=${OPTARG} ;;
            r) local redirect=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    echo "${code} ${cid} ${secret} ${redirect}"

    [ -z "${code+x}" ] && echo ${usage} && return 1
    [ -z "${cid+x}" ] && echo ${usage} && return 1
    [ -z "${secret+x}" ] && echo ${usage} && return 1
    [ -z "${redirect+x}" ] && echo ${usage} && return 1

    curl \
        -XGET \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
    "https://file-server:9877/token?grant_type=authorization_code&client_id=${cid}&client_secret=${secret}&scope=read&code=${code//[$'\t\r\n ']}&redirect_uri=${redirect}"
}
