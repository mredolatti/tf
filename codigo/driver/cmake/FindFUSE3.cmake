cmake_minimum_required(VERSION 3.14)

set(FUSE3_FOUND TRUE)
set(FUSE3_LIBRARIES )
set(FUSE3_DEFINITIONS )
set(FUSE3_INCLUDE_DIRS )

find_package(PkgConfig)

set(PC_FUSE3_INCLUDE_DIRS)
set(PC_FUSE3_LIBRARY_DIRS)
if(PKG_CONFIG_FOUND)
    pkg_check_modules(PC_FUSE3 "fuse3" QUIET)
    if(PC_FUSE3_FOUND)
        set(FUSE3_DEFINITIONS "${PC_FUSE3_CFLAGS_OTHER}")
    endif()
endif()

find_path(
    FUSE3_INCLUDE_DIRS
    NAMES fuse.h
    PATHS "${PC_FUSE3_INCLUDE_DIRS}"
    DOC "Include directories for FUSE3"
)

if(NOT FUSE3_INCLUDE_DIRS)
    set(FUSE3_FOUND FALSE)
endif()

find_library(
    FUSE3_LIBRARIES
    NAMES "fuse3"
    PATHS "${PC_FUSE3_LIBRARY_DIRS}"
    DOC "Libraries for FUSE3"
)

if(NOT FUSE3_LIBRARIES)
    set(FUSE3_FOUND FALSE)
endif()

if(FUSE3_FOUND)
    if(EXISTS "${FUSE3_INCLUDE_DIRS}/fuse3/fuse_common.h")
        file(READ "${FUSE3_INCLUDE_DIRS}/fuse3/fuse_common.h" _contents)
        string(REGEX REPLACE ".*# *define *FUSE_MAJOR_VERSION *([0-9]+).*" "\\1" FUSE3_MAJOR_VERSION "${_contents}")
        string(REGEX REPLACE ".*# *define *FUSE_MINOR_VERSION *([0-9]+).*" "\\1" FUSE3_MINOR_VERSION "${_contents}")
        set(FUSE3_VERSION "${FUSE3_MAJOR_VERSION}.${FUSE3_MINOR_VERSION}")
    endif()
endif()

if(FUSE3_INCLUDE_DIRS)
    include(FindPackageHandleStandardArgs)
    if(FUSE3_FIND_REQUIRED AND NOT FUSE3_FIND_QUIETLY)
        find_package_handle_standard_args(FUSE3 REQUIRED_VARS FUSE3_LIBRARIES FUSE3_INCLUDE_DIRS VERSION_VAR FUSE3_VERSION)
    else()
        find_package_handle_standard_args(FUSE3 "FUSE3 not found" FUSE3_LIBRARIES FUSE3_INCLUDE_DIRS)
    endif()
else(FUSE3_INCLUDE_DIRS)
    if(FUSE3_FIND_REQUIRED AND NOT FUSE3_FIND_QUIETLY)
        message(FATAL_ERROR "Could not find FUSE3 include directory")
    endif()
endif()


mark_as_advanced(
    FUSE3_INCLUDE_DIRS
    FUSE3_LIBRARIES
    FUSE3_DEFINITIONS
)

if(FUSE3_FOUND)
    if(NOT TARGET FUSE3::libfuse)
        add_library(FUSE3::libfuse UNKNOWN IMPORTED)
        set_target_properties(FUSE3::libfuse PROPERTIES
            INTERFACE_INCLUDE_DIRECTORIES "${FUSE3_INCLUDE_DIRS}"
            IMPORTED_LOCATION "${FUSE3_LIBRARIES}"
            IMPORTED_LINK_INTERFACE_LANGUAGES "C"
        )
        target_compile_definitions(FUSE3::libfuse
            INTERFACE "${FUSE3_DEFINITIONS}"
        )
    endif()
endif() 
