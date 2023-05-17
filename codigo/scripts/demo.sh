#!/usr/bin/env bash

# importar funciones de apiclients
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/is_api.sh"
source "${SCRIPT_DIR}/fs_api.sh"


# ---- DESCOMENTAR ESTO PARA LA DEMO (ya esta probado y anda)

# crear una cuenta en index server, iniciar sesion, habilitar 2fa y listar mapeos
# is_signup -n "martin" -e "mredolatti@test.com" -p "qwerty"
# token=$(is_login -e "mredolatti@test.com" -p "qwerty" 2> /dev/null \
# 	| jq -M '.token' \
# 	| tr -d '"')
# is_list -t "${token}" 2> /dev/null; echo ""
# is_setup_2fa -t "${token}" -q qr_temp.png
# xdg-open qr_temp.png
# 
# echo "ingresar codigo 2fa:"
# read tfa_passcode
# token=$(is_login -e "mredolatti@test.com" -p "qwerty" -f "${tfa_passcode}" 2> /dev/null \
# 	| jq -M '.token' \
# 	| tr -d '"')
# 
# 
# # --- VOLAR ESTO PARA LA DEMO
token="yGre6HAyw2TpRmoaTImmTahxW6wgJjfIq4yh7LtvjtK7cfjbiwjWrv2l70CJXEPS1HQ="
#------

is_list -t "${token}"

# Agregar organizacion unicen (de donde provienen los servidores que seran registrados)
#is_admin_add_org -n unicen

# ---- Levantar File server ----

# Listar organizaciones y servers
org_id=$(is_list_orgs -t "${token}" 2> /dev/null \
    | jq '.[] | select(.name=="unicen").id' \
    | tr -d '"') 

server_id=$(is_list_servers -t "${token}" -o "${org_id}" 2> /dev/null \
    | jq ' .[] | select(.name=="file-server").id ' \
    | tr -d '"')

# Vincular cuenta en index-server a cuenta en file-server-1
is_link_fs -t "${token}" -i "${server_id}"
