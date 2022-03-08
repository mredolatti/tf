#!/usr/bin/env bash

function idx_list() {
    curl -XGET 'http://index-server:9876/main/mappings'
}

#function idx_get() {
#    usage="usage: fs_get -i <file_id>"
#    local OPTIND
#    while getopts "i:" options; do
#        case ${options} in
#            i) local fid=${OPTARG} ;;
#            h) echo ${usage} && return 0 ;;
#            *) echo ${usage} && return 1 ;;
#        esac
#    done
#
#    [ -z ${fid+x} ] && echo ${usage} && return 1
#
#    curl \
#        -XGET \
#        --cacert ${FS_CACERT} \
#        --cert ${FS_CERT} \
#        --key ${FS_KEY} \
#        "https://file-server:9877/files/${fid}"
#}
