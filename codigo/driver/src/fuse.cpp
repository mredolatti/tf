#include "fuse.hpp"
#include "filemanager.hpp"
#define FUSE_USE_VERSION 35

#include <fuse3/fuse.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include <errno.h>
#include <iostream>


// helpers: mover a otro lado
mode_t dtype2stmode(unsigned char in)
{
    // las constantes son las mismas pero corridas unos bits para la izq
    // https://github.com/bminor/glibc/blob/master/io/sys/stat.h#L104
    // https://github.com/bminor/glibc/blob/master/sysdeps/unix/sysv/linux/bits/stat.h#L32
    // https://github.com/bminor/glibc/blob/master/dirent/dirent.h#L100
    return in << 12;
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

    SPDLOG_LOGGER_TRACE(ctx->logger(), "gathering statsa for: '{}'", path);
    auto res{ctx->file_manager().stat(path)};
    if (!res) {
        SPDLOG_LOGGER_ERROR(ctx->logger(), "failed to stat '{}'", path);
        return 1;
    }

    auto& de{(*res)};
    stbuf->st_mode = (de.is_directory() ? S_IFDIR : S_IFREG);
    stbuf->st_nlink = (de.is_directory() ? 2 : 1);
    return 0;


    //    struct stat {
    //           dev_t     st_dev;         /* ID of device containing file */
    //           ino_t     st_ino;         /* Inode number */
    //           mode_t    st_mode;        /* File type and mode */
    //           nlink_t   st_nlink;       /* Number of hard links */
    //           uid_t     st_uid;         /* User ID of owner */
    //           gid_t     st_gid;         /* Group ID of owner */
    //           dev_t     st_rdev;        /* Device ID (if special file) */
    //           off_t     st_size;        /* Total size, in bytes */
    //           blksize_t st_blksize;     /* Block size for filesystem I/O */
    //           blkcnt_t  st_blocks;      /* Number of 512B blocks allocated */

    //           /* Since Linux 2.6, the kernel supports nanosecond
    //              precision for the following timestamp fields.
    //              For the details before Linux 2.6, see NOTES. */

    //           struct timespec st_atim;  /* Time of last access */
    //           struct timespec st_mtim;  /* Time of last modification */
    //           struct timespec st_ctim;  /* Time of last status change */

    //       #define st_atime st_atim.tv_sec      /* Backward compatibility */
    //       #define st_mtime st_mtim.tv_sec
    //       #define st_ctime st_ctim.tv_sec
    //       };

    // TODO(mredolatti): llenar stbuf con las propiedades necesarias
    // - st_dev ?
    // - st_ino ?
    // - st_mode: S_IFREG o S_IFDIR, dependiendo de si es un prefix (folder) o file
    // - st_uid y st_gid leidos de config
    // - n_link: 1?
    // - st_rdev 0
    // - st_size: deberia venir del server
    // - st_blksize ?
    // - st_blocks ?
    // - st_atim = st_mtim = st_ctm = timestamp de ultima modificacion (del server)

    return 0;
}

static int mifs_access(const char *path, int mask)
{
    // TODO(mredolatti) Ver si en algun caso corresponde denegar acceso a un archivo/path
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
        std::cout << "fallo `fm.list`\n";
        return 1; // TODO(mredolatti): devolver codigo apropiado
    }

    std::cout << "llegaron " << (*entries).size() << " direntries\n";

    for (auto&& de : (*entries)) {
        std::cout << "agregando: " << de.name() << '\n';
        struct stat st;
        memset(&st, 0, sizeof(st));
        st.st_ino = 0;
        st.st_mode = (de.is_directory() ? S_IFDIR : S_IFREG);
        if (filler(buf, de.name().c_str(), &st, 0, static_cast<enum fuse_fill_dir_flags>(0))) {
            break;
        }
    }




/*
    auto dp{opendir(path)};
    if (dp == nullptr) {
        return -errno;
    }

    struct dirent *de;
    while ((de = readdir(dp)) != nullptr) {
        struct stat st;
        memset(&st, 0, sizeof(st));
        st.st_ino = de->d_ino;
        st.st_mode = dtype2stmode(de->d_type);
        if (filler(buf, de->d_name, &st, 0, static_cast<enum fuse_fill_dir_flags>(0))) {
            break;
        }
    }

    closedir(dp);
    */
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
    // TODO(mredolatti): volar esto?
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
    auto res{link(from, to)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_chmod(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    (void) fi;
    // TODO(mredolatti): volar esto?
    return 0;
}

static int mifs_chown(const char *path, uid_t uid, gid_t gid, struct fuse_file_info *fi)
{
    (void) fi;
    // TODO(mredolatti): volar esto?
    return 0;
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

static int mifs_open(const char *path, struct fuse_file_info *fi)
{
    auto res{open(path, fi->flags)};
    if (res == -1) {
        return -errno;
    }

    fi->fh = res;
    return 0;
}

static int mifs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto fd{(fi == nullptr) ? open(path, O_RDONLY) : fi->fh};
    if (fd == -1) {
        return -errno;
    }

    auto res{pread(fd, buf, size, offset)};
    if (res == -1) {
        res = -errno;
    }

    if(fi == nullptr) {
        close(fd);
    }

    return res;
}

static int mifs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto fd{(fi == nullptr) ? open(path, O_WRONLY) : fi->fh};
    if (fd == -1) {
        return -errno;
    }

    auto res{pwrite(fd, buf, size, offset)};
    if (res == -1) {
        res = -errno;
    }

    if(fi == nullptr) {
        close(fd);
    }

    return res;
}

static int mifs_statfs(const char *path, struct statvfs *stbuf)
{
    auto res{statvfs(path, stbuf)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_release(const char *path, struct fuse_file_info *fi)
{
    (void) path;
    close(fi->fh);
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

static int mifs_opendir(const char *dir, struct fuse_file_info *info)
{
    (void)info;
    return 0;
}

static const struct fuse_operations mifs_oper = {
    .getattr    = mifs_getattr,
    .mkdir      = mifs_mkdir,
    .unlink     = mifs_unlink,
    .rmdir      = mifs_rmdir,
    .rename     = mifs_rename,
    .link       = mifs_link,
    .chmod      = mifs_chmod,
    .chown      = mifs_chown,
    .truncate   = mifs_truncate,
    .open       = mifs_open,
    .read       = mifs_read,
    .write      = mifs_write,
    .statfs     = mifs_statfs,
    .release    = mifs_release,
    .fsync      = mifs_fsync,
    .opendir    = mifs_opendir,
    .readdir    = mifs_readdir,
    .init       = mifs_init,
    .access     = mifs_access,
    .create     = mifs_create,
};

int init_fuse(int argc, char *argv[], ContextData& ctx)
{
    umask(0);
    return fuse_main(argc, argv, &mifs_oper, &ctx);
}

ContextData::ContextData(mifs::log::logger_t logger, mifs::FileManager& fm) :
    logger_{std::move(logger)},
    fm_{fm}
{}

mifs::log::logger_t& ContextData::logger() { return logger_; }
mifs::FileManager& ContextData::file_manager() { return fm_; }

