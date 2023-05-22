#include "fuse.hpp"
#include "filemanager.hpp"
#define FUSE_USE_VERSION 35

#include <dirent.h>
#include <errno.h>
#include <fuse3/fuse.h>
#include <iostream>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

namespace helpers
{
int get_st_nlink(const mifs::fstree::views::Wrapper& w);
__mode_t get_st_mode(const mifs::fstree::views::Wrapper& w);
int map_filemanager_error(mifs::FileManager::Error e);
} // namespace helpers

// fuse

static void *mifs_init(struct fuse_conn_info *conn, struct fuse_config *cfg)
{
    (void)conn;
    return fuse_get_context()->private_data;
}

static int mifs_getattr(const char *path, struct stat *stbuf, struct fuse_file_info *fi)
{
    (void)fi;
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    if (strcmp(path, "/") == 0) {
        stbuf->st_mode = S_IFDIR | 0755;
        stbuf->st_nlink = 2;
        return 0;
    }

    SPDLOG_LOGGER_TRACE(ctx->logger(), "gathering stats for: '{}'", path);
    auto res{ctx->file_manager().stat(path)};
    if (!res) {
        SPDLOG_LOGGER_ERROR(ctx->logger(), "failed to stat '{}': {}", path, static_cast<int>(res.error()));
        return helpers::map_filemanager_error(res.error());
    }

    auto& de{(*res)};
    stbuf->st_mode = helpers::get_st_mode(de);
    stbuf->st_nlink = helpers::get_st_nlink(de);
    stbuf->st_size = de.size_bytes();
    stbuf->st_uid = getuid();
    stbuf->st_gid = getgid();
    stbuf->st_mtim.tv_sec = de.last_updated_seconds();
    return 0;
}

static int mifs_access(const char *path, int mask) { return 0; }

static int mifs_readdir(const char *path, void *buf, fuse_fill_dir_t filler, off_t offset,
                        struct fuse_file_info *fi, enum fuse_readdir_flags flags)
{

    (void)offset;
    (void)fi;
    (void)flags;

    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    auto entries{ctx->file_manager().list(path)};
    if (!entries) {
        return -ENOENT;
    }

    for (auto&& de : (*entries)) {
        struct stat st;
        memset(&st, 0, sizeof(st));
        st.st_ino = 0;
        st.st_size = de.size_bytes();
        st.st_mode = helpers::get_st_mode(de);
        st.st_mtim.tv_sec = de.last_updated_seconds();

        if (filler(buf, de.name().c_str(), &st, 0, static_cast<enum fuse_fill_dir_flags>(0))) {
            break;
        }
    }
    return 0;
}

static int mifs_mkdir(const char *path, mode_t mode)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().mkdir(path));
}

static int mifs_unlink(const char *path)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().remove(path));
}

static int mifs_rmdir(const char *path)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().rmdir(path));
}

static int mifs_symlink(const char *from, const char *to)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().link(from, to));
}

static int mifs_rename(const char *from, const char *to, unsigned int flags)
{
    (void)flags;
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().rename(from, to));
}

static int mifs_truncate(const char *path, off_t size, struct fuse_file_info *fi)
{
    // TODO(mredolatti): implementar!
    return 0;
}

static int mifs_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().touch(path));
}

static int mifs_readlink(const char *path, char *buffer, size_t buffer_size)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_TRACE(ctx->logger(), "reading link for: '{}'", path);
    auto res{ctx->file_manager().stat(path)};
    if (!res) {
        return helpers::map_filemanager_error(res.error());
    }

    const auto *as_link{res->link()};
    assert(as_link);

    auto resolved{fmt::format("{}/servers/{}/{}/{}", ctx->mount_point(), as_link->organization_name,
                              as_link->server_name, as_link->ref)};
    std::size_t idx{};
    while (buffer_size > 1 && idx < resolved.size()) {
        buffer[idx] = resolved[idx];
        idx++;
        buffer_size--;
    }
    buffer[idx] = '\0';

    return 0;
}

static int mifs_open(const char *path, struct fuse_file_info *fi)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "opening file: {}", path);
    fi->fh = ctx->file_manager().open(path, 0 /*TODO */);

    // TODO(mredolatti): error out if directory does not exist
    return 0;
}

static int mifs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "reading from file: {}", path);

    std::size_t read_bytes;
    mifs::FileManager::Error err;
    if (fi == nullptr) {
        SPDLOG_LOGGER_TRACE(ctx->logger(), "reading by path: {}", path);
        std::tie(read_bytes, err) = ctx->file_manager().read(path, buf, offset, size);
    } else {
        SPDLOG_LOGGER_TRACE(ctx->logger(), "reading by fd: {}", fi->fh);
        std::tie(read_bytes, err) = ctx->file_manager().read(fi->fh, buf, offset, size);
    }
    return err == mifs::FileManager::Error::Ok ? read_bytes : helpers::map_filemanager_error(err);
}

static int mifs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{

    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "writing to file: {}", path);
    auto [written_bytes, err]{ctx->file_manager().write(path, buf, size, offset)};
    return err == mifs::FileManager::Error::Ok ? written_bytes : helpers::map_filemanager_error(err);
}

static int mifs_fsync(const char *path, int isdatasync, struct fuse_file_info *fi)
{
    (void)path;
    (void)isdatasync;
    (void)fi;
    return 0;
}

static int mifs_opendir(const char *dir, struct fuse_file_info *fi)
{
    (void)fi;
    return 0;
}

static int mifs_flush(const char *path, struct fuse_file_info *fi)
{
    (void)fi;
    auto ctx{reinterpret_cast<ContextData *>(fuse_get_context()->private_data)};
    return helpers::map_filemanager_error(ctx->file_manager().flush(path));
}

static int mifs_release(const char *path, struct fuse_file_info *fi)
{
    (void)path;
    close(fi->fh);
    return 0;
}

static int mifs_utimens(const char *, const struct timespec tv[2], struct fuse_file_info *fi) { return 0; }

static const struct fuse_operations mifs_oper = {
    .getattr = mifs_getattr,
    .readlink = mifs_readlink,
    .mkdir = mifs_mkdir,
    .unlink = mifs_unlink,
    .rmdir = mifs_rmdir,
    .symlink = mifs_symlink,
    .rename = mifs_rename,
    .truncate = mifs_truncate,
    .open = mifs_open,
    .read = mifs_read,
    .write = mifs_write,
    .flush = mifs_flush,
    .release = mifs_release,
    .fsync = mifs_fsync,
    .opendir = mifs_opendir,
    .readdir = mifs_readdir,
    .init = mifs_init,
    .access = mifs_access,
    .create = mifs_create,
    .utimens = mifs_utimens,
};

// --------------------------------------------------

int init_fuse(int argc, char *argv[], ContextData& ctx)
{
    umask(0);
    return fuse_main(argc, argv, &mifs_oper, &ctx);
}

ContextData::ContextData(std::string mount_point, mifs::log::logger_t logger, mifs::FileManager& fm)
    : logger_{std::move(logger)},
      fm_{fm}
{
    char mount_point_buffer[1024];
    realpath(mount_point.c_str(), mount_point_buffer);
    mount_point_ = std::string{mount_point_buffer};
}

mifs::log::logger_t& ContextData::logger() { return logger_; }
mifs::FileManager& ContextData::file_manager() { return fm_; }
const std::string& ContextData::mount_point() const { return mount_point_; }

//---------------------------------------

namespace helpers
{

using mifs::fstree::views::Type;

__mode_t get_st_mode(const mifs::fstree::views::Wrapper& w)
{
    switch (w.type()) {
    case Type::File: return S_IFREG | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
    case Type::Folder: return S_IFDIR | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
    case Type::Link: return S_IFLNK | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
    }
    assert(false); // shold never get here
}

int get_st_nlink(const mifs::fstree::views::Wrapper& w)
{
    switch (w.type()) {
    case Type::File: return 1;
    case Type::Folder: return 2;
    case Type::Link: return 1;
    }
    assert(false); // shold never get here
}

int map_filemanager_error(mifs::FileManager::Error e)
{
    switch (e) {
    case mifs::FileManager::Error::Ok: return 0;
    case mifs::FileManager::Error::NotFound: return -ENOENT;
    case mifs::FileManager::Error::AlreadyExists: return -EEXIST;
    case mifs::FileManager::Error::NotAFile:
    case mifs::FileManager::Error::NotALink:
    case mifs::FileManager::Error::NotAForlder:
    case mifs::FileManager::Error::CannotWriteInNonServerPath:
    case mifs::FileManager::Error::ServerTreeManipulation:
    case mifs::FileManager::Error::InvalidLinkSource:
    case mifs::FileManager::Error::InvalidLinkDestination: return -EPERM;
    case mifs::FileManager::Error::FailedToFetchMappings:
    case mifs::FileManager::Error::FiledToUpdateRemoteMapping:
    case mifs::FileManager::Error::FiledToReadFileFromServer:
    case mifs::FileManager::Error::FiledToWriteFileInServer:
    case mifs::FileManager::Error::FailedToFetchServerInfos: return -EBADE;
    case mifs::FileManager::Error::InternalCacheError:
    case mifs::FileManager::Error::InternalRepresentationError: return EBADFD;
    case mifs::FileManager::Error::Unknown: break;
    }

    return -EPROTO;
}

} // namespace helpers
