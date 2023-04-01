# Incluir el finder de fuse3
list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")
find_package(FUSE3 REQUIRED)
find_package(CURL REQUIRED)
find_package(RapidJSON REQUIRED)
find_package(fmt REQUIRED)
find_package(spdlog REQUIRED)

# Ejecutable principal
add_executable(mifs)
target_sources(mifs
    PRIVATE main.cpp
    PRIVATE fuse.cpp
    PRIVATE http.cpp
    PRIVATE httpc.cpp
    PRIVATE tls.cpp
    PRIVATE isclient.cpp
    PRIVATE fsclient.cpp
    PRIVATE mappings.cpp
    PRIVATE filemeta.cpp
    PRIVATE jsend.cpp
    PRIVATE filemanager.cpp
    PRIVATE log.cpp
    PRIVATE fsmirror.cpp
    PRIVATE fselems.cpp
    PRIVATE fstree.cpp
    PRIVATE filecache.cpp
    PRIVATE openfiles.cpp
    PRIVATE servers.cpp
    PRIVATE istokens.cpp
    PRIVATE config.cpp
    PRIVATE fuseopts.cpp
    PRIVATE fscatalog.cpp
)

target_link_libraries(mifs
    PRIVATE FUSE3::libfuse
    PRIVATE CURL::libcurl
    PRIVATE RapidJSON::RapidJSON
    PRIVATE fmt::fmt
    PRIVATE spdlog::spdlog
)

set_target_properties(mifs PROPERTIES
    CXX_STANDARD 17
)
