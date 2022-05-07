#!/usr/bin/env bash

FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
FS_CERT=${FS_CERT:-'PKI/client/certs/client.crt'}
FS_KEY=${FS_KEY:-'PKI/client/private/client.key'}


function idx_list() {
    curl -XGET 'http://index-server:9876/mappings'
}

function idx_link_fs() {
    usage="usage: idx_link_fs -s <server_id>"
   local OPTIND
   while getopts "s:" options; do
       case ${options} in
           s) local sid=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${sid+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XGET \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       "http://index-server:9876/accounts/server/${sid}/authorize"
}
