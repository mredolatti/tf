#!/user/bin/env bash

# Initialize defaults if not set by user
[[ -z "${MONGO_CONTAINER}" ]] && MONGO_CONTAINER='mongomifs'
[[ -z "${MONGO_USER}" ]] && MONGO_USER=''
[[ -z "${MONGO_PASS}" ]] && MONGO_PASS=''
[[ -z "${MONGO_DB_INDEX}" ]] && MONGO_DB_INDEX='mifs_indexsrv'
[[ -z "${MONGO_DB_FILE}" ]] && MONGO_DB_FILE='mifs_filesrv'
[[ -z "${MONGO_HOST}" ]] && MONGO_HOST='localhost'

function mongo_query() {
    db="${1}"
    query="${2}"
    res=$(docker exec -i "${MONGO_CONTAINER}" mongosh "${db}" --quiet --eval "${query}")
    echo "${res}"
}

function mongo_insert_one() {
    db="${1}"
    collection="${2}"
    document="${3}"
    res=$(mongo_query "${db}" "db.${collection}.insertOne(${document})")
    echo "${res}" | awk '/insertedId/ { match($0, / insertedId: ObjectId\("([^"]+)"\)/, arr); print arr[1] }'
}


# Fetch the latest image if necessary, otherwhise do nothing.
# Return erroneous exit code if image fetch fails
function fetch_image() {
    images=$(docker images | grep "mongo" | wc -l)
    if [[ "$images" == "0" ]]; then
        docker pull mongo:latest
        return $?
    fi
}

function drop_db() {
    mongo_query "${MONGO_DB_INDEX}" "printjson(db.dropDatabase())"
    mongo_query "${MONGO_DB_FILE}" "printjson(db.dropDatabase())"
}

function init_db() {
    # Index-server application database index setup
    mongo_query ${MONGO_DB_INDEX} "printjson(db.Organizations.createIndex({name: 1}, {unique: true}))"
    mongo_query ${MONGO_DB_INDEX} "printjson(db.UserAccounts.createIndex({userId: 1}, {unique: false}))"
    mongo_query ${MONGO_DB_INDEX} "printjson(db.UserAccounts.createIndex({userId: 1, organizationName: 1, serverName: 1}, {unique: true}))"
    mongo_query ${MONGO_DB_INDEX} "printjson(db.FileServers.createIndex({organizationName: 1, name: 1}, {unique: true}))"
    mongo_query ${MONGO_DB_INDEX} "printjson(db.Mappings.createIndex({userId: 1, path: 1}, {unique: true}))"
    mongo_query ${MONGO_DB_INDEX} "printjson(db.PendingOAuth2.createIndex({state: 1}, {unique: false}))"

    # TODO(mredolatti): setup indexes for file-server app
}

function setup_fixtures() {
    org1=$(mongo_insert_one "${MONGO_DB_INDEX}" "Organizations" "$(f_org 'organization1')")

    fs1=$(mongo_insert_one \
        ${MONGO_DB_INDEX} \
        "FileServers" \
        "$(f_fs 'org1' 'fs1' 'servercito' 'https://file-server:9877/authorize' 'https://file-server:9877/token' 'https://file-server:9877/file' 'file-server:9000')" \
    )

    passhash=$(htpasswd  -B -C10 -nb "martinredolatti@gmail.com" "123456" | cut -d':' -f2 | awk NF)
    user1=$(mongo_insert_one "${MONGO_DB_INDEX}" "Users" "$(f_user_id '63ad6d1c01c2a1a5c1259b9f' 'Martín Redolatti' 'martinredolatti@gmail.com' ${passhash})")

    _=$(mongo_insert_one "${MONGO_DB_INDEX}" "Mappings" "$(f_map ${user1} 'org1' 'fs1' 'file1.txt' '0' 'path/to/file1.txt' '1646394925714181390')")
    _=$(mongo_insert_one "${MONGO_DB_INDEX}" "Mappings" "$(f_map ${user1} 'org1' 'fs1' 'file2.txt' '0' 'path/to/file2.txt' '1646394925714181390')")
    _=$(mongo_insert_one "${MONGO_DB_INDEX}" "Mappings" "$(f_map ${user1} 'org1' 'fs1' 'file3.txt' '0' 'path/to/another/file3.txt' '1646394925714181390')")
    _=$(mongo_insert_one "${MONGO_DB_INDEX}" "Mappings" "$(f_map ${user1} 'org1' 'fs1' 'file4.txt' '0' 'path/to/another/file4.txt' '1646394925714181390')")
    _=$(mongo_insert_one "${MONGO_DB_INDEX}" "Mappings" "$(f_map ${user1} 'org1' 'fs1' 'file5.txt' '0' 'path/to/yet/another/file5.txt' '1646394925714181390')")

}

function start_shell() {
    docker exec -it ${MONGO_CONTAINER} mongosh
}

function usage() {
      echo "$0 [-f]"
      echo " -f drop the database prior to attempting creation"
      exit 0
}

# ----------
# Formatter helper functions
# @{
function f_org() {
    echo "{name: '${1}'}"
}

function f_fs() {
    echo "{organizationName: '${1}', name: '${2}', authUrl: '${3}', tokenUrl: '${4}', fetchUrl: '${5}', controlEndpoint: '${6}'}"
}

function f_user() {
    echo "{name: '${1}', email: '${2}', accessToken: '${3}', refreshToken: '${4}'}"
}

function f_user_id() {
	# 
    echo "{_id: ObjectId('${1}'), name: '${2}', email: '${3}', password: '${4}'}"
}

function f_map() {
    echo "{userId: ObjectId('${1}'), organizationName: '${2}', serverName: '${3}', ref: '${4}', sizeBytes: ${5}, path: '${6}', updated: ${7}}"
}
# @}


# ----------
# Main program execution

# Parse CLI args
while getopts ":hbdcfs" opt; do
    case $opt in
        h)
            usage
            ;;
        b)
            build=1
            ;;
        d)
            drop_db=1
            ;;
        c)
            create=1
            ;;
        f)
            fixtures=1
            ;;
        s)
            shell=1

    esac
done

if [[ ! -z "${build}" ]]; then # Download image and create container
    fetch_image
    create_container
fi

if [[ ! -z "${drop_db}" ]]; then # Drop database if requested
    drop_db
fi

if [[ ! -z "${create}" ]]; then # Create database, extensions & tables
    init_db
fi

if [[ ! -z "${fixtures}" ]]; then # Insert fixtures/test data if requested
    setup_fixtures
fi

if [[ ! -z "${shell}" ]]; then # Open a shell
    start_shell
fi
