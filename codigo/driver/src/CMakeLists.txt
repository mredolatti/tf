
# Libreria
list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake") # Incluir el finder de fuse3
find_package(FUSE3 REQUIRED)
find_package(CURL REQUIRED)
find_package(RapidJSON REQUIRED)
find_package(fmt REQUIRED)
find_package(spdlog REQUIRED)

add_library(mifs STATIC)
set_target_properties(mifs PROPERTIES CXX_STANDARD 20)

target_sources(mifs
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
    PRIVATE fstree.cpp
    PRIVATE filecache.cpp
    PRIVATE openfiles.cpp
    PRIVATE servers.cpp
    PRIVATE istokens.cpp
    PRIVATE config.cpp
    PRIVATE fscatalog.cpp
    PRIVATE apierror.cpp
    PRIVATE nsresp.cpp)

target_link_libraries(mifs
    PRIVATE FUSE3::libfuse
    PRIVATE CURL::libcurl
    PRIVATE RapidJSON::RapidJSON
    PRIVATE fmt::fmt
    PRIVATE spdlog::spdlog)

# ----------------------------

# Driver
add_executable(mount.mifs)
set_target_properties(mount.mifs PROPERTIES CXX_STANDARD 20)

target_sources(mount.mifs
    PRIVATE main.cpp
    PRIVATE fuse.cpp
    PRIVATE fuseopts.cpp)

target_link_libraries(mount.mifs mifs)

# ----------------------------

# Tools
find_package(ImageMagick COMPONENTS Magick++)

add_executable(mifs-tools)
set_target_properties(mifs-tools PROPERTIES CXX_STANDARD 20)

target_sources(mifs-tools
    PRIVATE tools.cpp)

target_link_libraries(mifs-tools
    PRIVATE mifs
    PRIVATE ${ImageMagick_LIBRARIES})

target_include_directories(mifs-tools
    PRIVATE ${ImageMagick_INCLUDE_DIRS})

