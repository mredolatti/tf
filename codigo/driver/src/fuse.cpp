#include "fuse.hpp"
#include "filemanager.hpp"
#define FUSE_USE_VERSION 35

#include <fuse3/fuse.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include <errno.h>
#include <iostream>

class EntryInfo : public mifs::types::FSElemVisitor
{
    public:
    void visit_file(const mifs::types::FSEFile&) override;
    void visit_link(const mifs::types::FSELink&) override;
    void visit_folder(const mifs::types::FSEFolder&) override;
    
    __mode_t st_mode() const;
    int st_nlink() const;

    private:
    enum class Type
    {
        File,
        Folder,
        Link
    };
    Type type_;
};

void EntryInfo::visit_file(const mifs::types::FSEFile&)     { type_ = Type::File; }
void EntryInfo::visit_link(const mifs::types::FSELink&)     { type_ = Type::Link; }
void EntryInfo::visit_folder(const mifs::types::FSEFolder&) { type_ = Type::Folder; }

__mode_t EntryInfo::st_mode() const
{
    switch (type_) {
        case Type::File:    return S_IFREG | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
        case Type::Folder:  return S_IFDIR | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
        case Type::Link:    return S_IFLNK | S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP;
    }
    assert(false); // shold never get here
}

int EntryInfo::st_nlink() const
{
    switch (type_) {
        case Type::File:    return 1;
        case Type::Folder:  return 2;
        case Type::Link:    return 1;
    }
    assert(false); // shold never get here
}

// fuse

static void* mifs_init(struct fuse_conn_info *conn, struct fuse_config *cfg)
{
    (void) conn;
    return fuse_get_context()->private_data;
}

static int mifs_getattr(const char* path, struct stat* stbuf, struct fuse_file_info* fi)
{
    (void) fi;
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    if (strcmp(path, "/") == 0) {
        stbuf->st_mode = S_IFDIR | 0755;
        stbuf->st_nlink = 2;
        return 0;
    }

    SPDLOG_LOGGER_TRACE(ctx->logger(), "gathering stats for: '{}'", path);
    auto res{ctx->file_manager().stat(path)};
    if (!res) {
        SPDLOG_LOGGER_ERROR(ctx->logger(), "failed to stat '{}': {}", path, res.error());
        return -ENOENT;
    }

    auto& de{(*res)};
    EntryInfo v;
    de->accept(v);
    stbuf->st_mode = v.st_mode();
    stbuf->st_nlink = v.st_nlink();
    stbuf->st_size = de->size_bytes();
    stbuf->st_uid = getuid();
    stbuf->st_gid = getgid();

    if (const auto* as_file{dynamic_cast<mifs::types::FSEFile*>(de.get())}) {
        stbuf->st_mtim.tv_sec = as_file->last_updated();
    }

    return 0;
}

static int mifs_access(const char *path, int mask)
{
    return 0;
}

static int mifs_readdir(
    const char *path, 
    void *buf, 
    fuse_fill_dir_t filler, 
    off_t offset, 
    struct fuse_file_info *fi,
    enum fuse_readdir_flags flags
) {

    (void) offset;
    (void) fi;
    (void) flags;

    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    auto entries{ctx->file_manager().list(path)};
    if (!entries) {
        return 1; // TODO(mredolatti): devolver codigo apropiado
    }

    for (auto&& de : (*entries)) {
        if (!de) {
            SPDLOG_LOGGER_ERROR(ctx->logger(), "null direntry");
            std::abort();
        }

        SPDLOG_LOGGER_INFO(ctx->logger(), "agregando direntry: {}", de->name());
        EntryInfo v;
        de->accept(v);
        struct stat st;
        memset(&st, 0, sizeof(st));
        st.st_ino = 0;
        st.st_mode = v.st_mode();

	if (const auto* as_file{dynamic_cast<mifs::types::FSEFile*>(de.get())}) {
		st.st_mtim.tv_sec = as_file->last_updated();
	}

        if (filler(buf, de->name().c_str(), &st, 0, static_cast<enum fuse_fill_dir_flags>(0))) {
            break;
        }
    }
    return 0;
}

static int mifs_mkdir(const char *path, mode_t mode)
{
    // TODO(mredolatti): ver como manejar esto correctamente:
    // - creacion de dirs solo en memoria ?
    // - no hacer nada ?
    return 0;
}

static int mifs_unlink(const char *path)
{
    // TODO(mredolatti): que hacer aca? volver a mover el archivo a su ubicacion original /<server>/file_Ref?
    return 0;
}

static int mifs_rmdir(const char *path)
{
    // TODO(mredolatti): borrado logico en memoria?
    return 0;
}

static int mifs_symlink(const char *from, const char *to)
{
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    ctx->file_manager().link(from, to);
    return 0;
}

static int mifs_rename(const char *from, const char *to, unsigned int flags)
{
    if (flags) {
        return -EINVAL;
    }

    // TODO(mredolatti): hacer un rename del mapping en index-server
    return 0;
}

static int mifs_link(const char *from, const char *to)
{
    return EACCES; // nohard links allowed
}

static int mifs_chmod(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    return EACCES;
}

static int mifs_chown(const char *path, uid_t uid, gid_t gid, struct fuse_file_info *fi)
{
    return EACCES;
}

static int mifs_truncate(const char *path, off_t size,
            struct fuse_file_info *fi)
{
    // TODO(mredolatti): volar esto?
    return 0;
}

static int mifs_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    auto res{open(path, fi->flags, mode)};
    if (res == -1) {
        return -errno;
    }

    fi->fh = res;
    return 0;
}

static int mifs_readlink(const char* path, char* buffer, size_t buffer_size)
{
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_TRACE(ctx->logger(), "reading link for: '{}'", path);
    auto res{ctx->file_manager().stat(path)};
    if (!res) {
        return 1; // TODO
    }

    const auto* as_link{dynamic_cast<const mifs::types::FSELink*>((*res).get())};
    assert(as_link);

    auto resolved{fmt::format("{}/servers/{}/{}/{}", ctx->mount_point(), as_link->org_name(), as_link->server_name(), as_link->ref())};
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
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "opening file: {}", path);
    fi->fh = ctx->file_manager().open(path, 0 /*TODO */);
    return 0;
}

static int mifs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "reading from file: {}", path);

    std::size_t read_bytes;
    if (fi == nullptr) {
        SPDLOG_LOGGER_TRACE(ctx->logger(), "reading by path: {}", path);
        read_bytes = ctx->file_manager().read(path, buf, offset, size);
    } else {
        SPDLOG_LOGGER_TRACE(ctx->logger(), "reading by fd: {}", fi->fh);
        read_bytes = ctx->file_manager().read(fi->fh, buf, offset, size);
    }
    return read_bytes;
}

static int mifs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{

    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    SPDLOG_LOGGER_INFO(ctx->logger(), "writing to file: {}", path);
    return ctx->file_manager().write(path, buf, size, offset);
}

static int mifs_statfs(const char *path, struct statvfs *stbuf)
{
    auto res{statvfs(path, stbuf)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_fsync(const char *path, int isdatasync,
             struct fuse_file_info *fi)
{
    (void) path;
    (void) isdatasync;
    (void) fi;
    return 0;
}

static int mifs_opendir(const char *dir, struct fuse_file_info* fi)
{
    (void)fi;
    return 0;
}

static int mifs_flush(const char* path, struct fuse_file_info* fi)
{
    (void)fi;
    auto ctx{reinterpret_cast<ContextData*>(fuse_get_context()->private_data)};
    ctx->file_manager().flush(path);
    return 0;
}

static int mifs_release(const char *path, struct fuse_file_info* fi)
{
    (void) path;
    close(fi->fh);
    return 0;
}


static const struct fuse_operations mifs_oper = {
    .getattr    = mifs_getattr,
    .readlink   = mifs_readlink,
    .mkdir      = mifs_mkdir,
    .unlink     = mifs_unlink,
    .rmdir      = mifs_rmdir,
    .symlink    = mifs_symlink,
    .rename     = mifs_rename,
    .link       = mifs_link,
    .chmod      = mifs_chmod,
    .chown      = mifs_chown,
    .truncate   = mifs_truncate,
    .open       = mifs_open,
    .read       = mifs_read,
    .write      = mifs_write,
    .statfs     = mifs_statfs,
    .flush      = mifs_flush,
    .release    = mifs_release,
    .fsync      = mifs_fsync,
    .opendir    = mifs_opendir,
    .readdir    = mifs_readdir,
    .init       = mifs_init,
    .access     = mifs_access,
    .create     = mifs_create,
};


// --------------------------------------------------

int init_fuse(int argc, char *argv[], ContextData& ctx)
{
    umask(0);
    return fuse_main(argc, argv, &mifs_oper, &ctx);
}

ContextData::ContextData(std::string mount_point, mifs::log::logger_t logger, mifs::FileManager& fm) :
    logger_{std::move(logger)},
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


