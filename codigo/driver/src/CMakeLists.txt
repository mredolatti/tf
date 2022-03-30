list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")
find_package(FUSE3 REQUIRED)

add_executable(mifs)
target_sources(mifs
    PRIVATE main.cpp
    PRIVATE fuse.cpp
)

target_link_libraries(mifs
    PRIVATE FUSE3::libfuse
)
