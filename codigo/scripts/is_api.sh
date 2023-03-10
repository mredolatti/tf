#!/usr/bin/env bash

FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
FS_CERT=${FS_CERT:-'PKI/client/certs/client.crt'}
FS_KEY=${FS_KEY:-'PKI/client/private/client.key'}


function is_signup() {
    usage="usage: is_signup -n <name> -e <email> -p <password>"
   local OPTIND
   while getopts "hn:e:p:" options; do
       case ${options} in
           n) local name=${OPTARG} ;;
           e) local email=${OPTARG} ;;
           p) local password=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${name+x} ] && echo ${usage} && return 1
   [ -z ${email+x} ] && echo ${usage} && return 1
   [ -z ${password+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XPOST \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       -H'Content-Type: application/json' \
       -d"{\"name\": \"${name}\", \"email\": \"${email}\", \"password\": \"${password}\"}" \
       "http://index-server:9876/signup"
}

function is_login() {
    usage="usage: is_login -e <email> -p <password> [-f <2fa_passcode>]"
   local OPTIND
   while getopts "he:p:f:" options; do
       case ${options} in
           e) local email=${OPTARG} ;;
           p) local password=${OPTARG} ;;
	   f) local passcode=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${email+x} ] && echo ${usage} && return 1
   [ -z ${password+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XPOST \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       -H'Content-Type: application/json' \
       -d"{\"email\": \"${email}\", \"password\": \"${password}\", \"OTP\": \"${passcode}\"}" \
       "http://index-server:9876/login"
}

function is_setup_2fa() {
   usage="usage: is_setup_2fa -t <session_token> -q <target_qr_code_filename>"
   local OPTIND
   while getopts "ht:q:" options; do
       case ${options} in
           t) local token=${OPTARG} ;;
	   q) local qr_output=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${token+x} ] && echo ${usage} && return 1
   [ -z ${qr_output+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XPOST \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       -H'Content-Type: application/json' \
       -H"X-MIFS-IS-Session-Token: ${token}" \
       "http://index-server:9876/2fa" \
       --output "${qr_output}"
}

function is_list() {
   usage="usage: is_list -t <session_token>"
   local OPTIND
   while getopts "ht:" options; do
       case ${options} in
           t) local token=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${token+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XGET \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       -H'Content-Type: application/json' \
       -H"X-MIFS-IS-Session-Token: ${token}" \
       "http://index-server:9876/mappings"
}

function is_link_fs() {
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

function is_logout() {
    usage="usage: is_logout -t <session_token>"
   local OPTIND
   while getopts "ht:" options; do
       case ${options} in
           t) local token=${OPTARG} ;;
           h) echo ${usage} && return 0 ;;
           *) echo ${usage} && return 1 ;;
       esac
   done

   [ -z ${token+x} ] && echo ${usage} && return 1

   curl \
       -v \
       -L \
       -XPOST \
       --cacert ${FS_CACERT} \
       --cert ${FS_CERT} \
       --key ${FS_KEY} \
       -H'Content-Type: application/json' \
       -H"X-MIFS-IS-Session-Token: ${token}" \
       "http://index-server:9876/logout"
}


