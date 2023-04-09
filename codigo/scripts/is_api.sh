#!/usr/bin/env bash

# Setup base URLs
BASE_CLIENTS_URL="https://index-server:9876/api/clients/v1"
BASE_FILESERVERS_URL="https://index-server:9876/api/fileservers/v1"

# Setup user key & cert (+cacert for server validation) to authenticate calls to FS
USER_FS_CACERT=${USER_FS_CACERT:-'PKI/root/certs/ca.crt'}
USER_FS_CERT=${USER_FS_CERT:-'PKI/client/certs/client.crt'}
USER_FS_KEY=${USER_FS_KEY:-'PKI/client/private/client.key'}

# Setup file server cert & key (to call index server on behalf of fs)
FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
FS_CERT=${FS_CERT:-'PKI/fileserver/certs/chain.pem'}
FS_CERT=${FS_CERT:-'PKI/fileserver/certs/chain.pem'}

function is_signup() {
    local usage="usage: is_signup -n <name> -e <email> -p <password> [-v]"
    local verbose=""
    local OPTIND
    while getopts "hvn:e:p:" options; do
        case ${options} in
            n) local name=${OPTARG} ;;
            e) local email=${OPTARG} ;;
            p) local password=${OPTARG} ;;
	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z ${name+x} ] && echo ${usage} && return 1
    [ -z ${email+x} ] && echo ${usage} && return 1
    [ -z ${password+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H'Content-Type: application/json' \
        -d"{\"name\": \"${name}\", \"email\": \"${email}\", \"password\": \"${password}\"}" \
        "${BASE_CLIENTS_URL}/signup"
}

function is_login() {
    local usage="usage: is_login -e <email> -p <password> [-f <2fa_passcode>]"
    local verbose=""
    local OPTIND
    while getopts "he:p:f:v" options; do
        case ${options} in
            e) local email=${OPTARG} ;;
            p) local password=${OPTARG} ;;
            f) local passcode=${OPTARG} ;;
	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${email+x} ] && echo ${usage} && return 1
    [ -z ${password+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H'Content-Type: application/json' \
        -d"{\"email\": \"${email}\", \"password\": \"${password}\", \"OTP\": \"${passcode}\"}" \
        "${BASE_CLIENTS_URL}/login"
}

function is_setup_2fa() {
    local usage="usage: is_setup_2fa -t <session_token> -q <target_qr_code_filename>"
    local verbose=""
    local OPTIND
    while getopts "ht:q:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            q) local qr_output=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done

    [ -z ${token+x} ] && echo ${usage} && return 1
    [ -z ${qr_output+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/2fa" \
        --output "${qr_output}"
}

function is_mapping_list() {
    local usage="usage: is_mapping_list -t <session_token>"
    local verbose=""
    local OPTIND
    while getopts "hvft:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            f) local force="?forceUpdate=true" ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1

    curl \
        ${verbose} \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/mappings${force}"
}

function is_mapping_rename() {
    local usage="usage: is_mapping_update -t <session_token> -i <mapping_id> -n <new_path>"
    local verbose=""
    local OPTIND
    while getopts "hvt:i:n:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            i) local id=${OPTARG} ;;
            n) local new=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPUT \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        -H"Content-Type: application/json" \
        -d"{\"path\": \"${new}\"}" \
        "${BASE_CLIENTS_URL}/mappings/${id}"
}

function is_mapping_delete() {
    local usage="usage: is_mapping_delete -t <session_token> -i <mapping_id>"
    local verbose=""
    local OPTIND
    while getopts "hvt:i:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            i) local id=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XDELETE \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/mappings/${id}"
}

function is_mapping_create() {
    local usage="usage: is_mapping_create -t <session_token> -o <org> -s <server> -r <ref> -p <path>"
    local verbose=""
    local OPTIND
    while getopts "hvt:o:s:r:p:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            o) local org=${OPTARG} ;;
            s) local server=${OPTARG} ;;
            r) local ref=${OPTARG} ;;
            p) local path=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
            v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        -H"Content-Type: application/json" \
        -d"{\"organizationName\": \"${org}\", \"serverName\": \"${server}\", \"ref\": \"${ref}\", \"path\": \"${path}\"}" \
        "${BASE_CLIENTS_URL}/mappings"
}

function is_list_orgs() {
    local usage="usage: is_list_orgs -t <session_token>"
    local verbose=""
    local OPTIND
    while getopts "hvt:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/organizations"
}

function is_list_servers_for_org() {
    local usage="usage: is_list_orgs -t <session_token> -o <org_name>"
    local verbose=""
    local OPTIND
    while getopts "hvt:o:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
	    o) local org=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
    [ -z ${org+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/organizations/${org}/servers"
}

function is_list_servers() {
    local usage="usage: is_list_servers -t <session_token>"
    local verbose=""
    local OPTIND
    while getopts "hvt:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
            h) echo ${usage} && return 0 ;;
	    v) verbose="-v" ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/servers"
}

function is_link_fs() {
    local usage="usage: idx_link_fs -o <org_name> -s <server_name> -t <token>"
    local verbose=""
    local OPTIND
    while getopts "hvt:o:s:" options; do
        case ${options} in
            s) local server=${OPTARG} ;;
            o) local org=${OPTARG} ;;
	    t) local token=${OPTARG} ;;
 	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
    [ -z ${org+x} ] && echo ${usage} && return 1
    [ -z ${server+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/organizations/${org}/servers/${server}/link"
}

function is_logout() {
    local usage="usage: is_logout -t <session_token>"
    local verbose=""
    local OPTIND
    while getopts "hvt:" options; do
        case ${options} in
            t) local token=${OPTARG} ;;
 	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${token+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        -H'Content-Type: application/json' \
        -H"X-MIFS-IS-Session-Token: ${token}" \
        "${BASE_CLIENTS_URL}/logout"
}


# ---- ADMIN API---------------

function is_admin_add_org() {
    local usage="usage: is_admin_add_org -n <org_name>"
    local verbose=""
    local OPTIND
    while getopts "hvn:" options; do
        case ${options} in
            n) local org_name=${OPTARG} ;;
 	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${org_name+x} ] && echo ${usage} && return 1
 
    curl \
        ${verbose} \
        -L \
        -XPOST \
        --cacert ${USER_FS_CACERT} \
	-H'Content-Type: application/json' \
	-d"{\"name\": \"${org_name}\"}" \
        "${BASE_CLIENTS_URL}/admin/organizations"
}


# -----FILE SERVERS API------------

function is_test() {
    curl \
        -v \
        -L \
        -XGET \
        --cacert ${USER_FS_CACERT} \
        --cert ${USER_FS_CERT} \
        --key ${USER_FS_KEY} \
        "${BASE_FILESERVERS_URL}/test"
}

function is_server_register() {
    local usage="usage: is_server_register -i <id> -n <name> -o <orgId> -a <authUrl> t <tokenUrl> -f <fetchUrl> -c <controlEndpoint>"
    local verbose=""
    local OPTIND
    while getopts "hvo:a:t:f:c:" options; do
        case ${options} in
	    o) local orgId=${OPTARG} ;;
	    a) local authUrl=${OPTARG} ;;
	    f) local fetchUrl=${OPTARG} ;;
	    c) local controlEndpoint=${OPTARG} ;;
            t) local tokenUrl=${OPTARG} ;;
 	    v) verbose="-v" ;;
            h) echo ${usage} && return 0 ;;
            *) echo ${usage} && return 1 ;;
        esac
    done
 
    [ -z ${orgId+x} ] && echo ${usage} && return 1
    [ -z ${authUrl+x} ] && echo ${usage} && return 1
    [ -z ${fetchUrl+x} ] && echo ${usage} && return 1
    [ -z ${controlEndpoint+x} ] && echo ${usage} && return 1
    [ -z ${tokenUrl+x} ] && echo ${usage} && return 1

    local body="{
    	\"orgId\": \"${orgId}\",
    	\"authUrl\": \"${authUrl}\",
    	\"fetchUrl\": \"${fetchUrl}\",
    	\"controlEndpoint\": \"${controlEndpoint}\",
    	\"tokenUrl\": \"${tokenUrl}\"
    }"

    curl \
        ${verbose} \
        -L \
        -XPOST \
        -H'Content-Type: application/json' \
	-d "${body}" \
        --cacert ${FS_CACERT} \
        --cert ${FS_CERT} \
        --key ${FS_KEY} \
        "${BASE_FILESERVERS_URL}/register"
}
